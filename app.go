package main

import (
	"container/heap"
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/hdt3213/rdb/parser"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// GetPrefixMulti 支持多种分隔符，返回前缀
func GetPrefixMulti(key string, level int) string {
	// 找到第一个分隔符
	seps := []string{":", "|", ".", "_", "-"}
	// 找到第一个分隔符
	sep := ""
	pos := len(key)
	for _, s := range seps {
		if i := strings.Index(key, s); i != -1 && i < pos {
			sep = s
			pos = i
		}
	}

	// 没有分隔符 → 只有 1 层
	if sep == "" {
		return ""
	}

	// 按分隔符切分
	parts := strings.Split(key, sep)
	if len(parts) < level {
		return "" // 层级不足
	}

	return strings.Join(parts[:level], sep)
}

func formatBytes(bytes int) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB",
		float64(bytes)/float64(div), "KMGTPE"[exp])
}

// RDB 分析结果的顶层结构
type RDBAnalysis struct {
	FileName          string `json:"file_name"` // 文件名
	TotalMemory       int    // 总内存占用 (字节)
	TotalMemoryReable string `json:"total_memory"`
	TotalKeys         int    `json:"total_keys"` // 总 key 数量

	// 各类型的统计
	TypeStats map[string]*TypeStat `json:"type_stats"`

	// 过期时间统计
	ExpireStats map[string]*TypeStat `json:"expire_stats"`

	// Top500 Key 按内存排序
	TopKeys    []*KeyInfo `json:"top_keys"`
	TopKesHeap *TopNKeys

	TopPrefix500Keys []*KeyInfo          `json:"top_prefix_keys"`
	PrefixMemMap     map[string]*KeyInfo // 前缀->内存
	PrefixTop500Heap *TopNKeys           // 前500个前缀
}

// 各类型统计（string, list, set, zset, hash）
type TypeStat struct {
	Count  int `json:"count"`  // key 数量
	Memory int `json:"memory"` // 内存占用
}

// 单个 Key 信息（主要用于 Top500）
type KeyInfo struct {
	DB          int    `json:"db"`   // 所属 DB
	Key         string `json:"key"`  // key 名称
	Type        string `json:"type"` // 类型
	Size        int    `json:"size"` // 内存大小
	ReadbleSize string `json:"readble_size"`
	Elements    int    `json:"elements"` //元素个数
	Expire      string `json:"expire"`   // 过期时间
}

// 前缀统计
type PrefixStat struct {
	Prefix string `json:"prefix"` // 前缀
	Count  int    `json:"count"`  // key 数量
	Memory uint64 `json:"memory"` // 占用内存
}

type KeyHeap []*KeyInfo

func (h KeyHeap) Len() int            { return len(h) }
func (h KeyHeap) Less(i, j int) bool  { return h[i].Size < h[j].Size } // 小顶堆
func (h KeyHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *KeyHeap) Push(x interface{}) { *h = append(*h, x.(*KeyInfo)) }
func (h *KeyHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

type TopNKeys struct {
	limit int
	heap  KeyHeap
}

func NewTopNKeys(limit int) *TopNKeys {
	return &TopNKeys{
		limit: limit,
		heap:  make(KeyHeap, 0, limit),
	}
}

func (t *TopNKeys) Add(info *KeyInfo) {
	if len(t.heap) < t.limit {
		heap.Push(&t.heap, info)
		return
	}
	if info.Size > t.heap[0].Size {
		t.heap[0] = info
		heap.Fix(&t.heap, 0)
	}
}

func (t *TopNKeys) Items() []*KeyInfo {
	result := make([]*KeyInfo, len(t.heap))
	copy(result, t.heap)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Size > result[j].Size // 降序
	})
	return result
}

func updateBasicStats(a *RDBAnalysis, o parser.RedisObject) {
	a.TotalKeys++
	a.TotalMemory += o.GetSize()
	switch o.GetType() {
	case "string":
		a.TypeStats["string"].Count++
		a.TypeStats["string"].Memory = a.TypeStats["string"].Memory + o.GetSize()
	case "set":
		a.TypeStats["set"].Count++
		a.TypeStats["set"].Memory = a.TypeStats["set"].Memory + o.GetSize()
	case "zset":
		a.TypeStats["zset"].Count++
		a.TypeStats["zset"].Memory = a.TypeStats["zset"].Memory + o.GetSize()
	case "list":
		a.TypeStats["list"].Count++
		a.TypeStats["list"].Memory = a.TypeStats["list"].Memory + o.GetSize()
	case "hash":
		a.TypeStats["hash"].Count++
		a.TypeStats["hash"].Memory = a.TypeStats["hash"].Memory + o.GetSize()
	}
	expire := ""
	//fmt.Println(a.TypeStats)
	if o.GetExpiration() == nil {
		expire = ""
	} else {
		expire = o.GetExpiration().Format("2006-01-02 15:04:05")
	}
	info := &KeyInfo{
		DB:          o.GetDBIndex(),
		Key:         o.GetKey(),
		Type:        o.GetType(),
		Size:        o.GetSize(),
		ReadbleSize: formatBytes(o.GetSize()),
		Elements:    o.GetElemCount(), // 如果支持
		Expire:      expire,           // 如果支持
	}

	a.TopKesHeap.Add(info)
}

func updateExpireStats(a *RDBAnalysis, o parser.RedisObject) {
	//fmt.Println(o.GetExpiration())
	if o.GetExpiration() == nil {
		a.ExpireStats["noexpire"].Count++
		a.ExpireStats["noexpire"].Memory = a.ExpireStats["noexpire"].Memory + o.GetSize()
	} else {
		duration := time.Since(*o.GetExpiration())
		switch h := duration.Hours(); {
		case h < 0:
			a.ExpireStats["expired"].Count++
			a.ExpireStats["expired"].Memory = a.ExpireStats["expired"].Memory + o.GetSize()
		case h <= 1 && h > 0:
			a.ExpireStats["exp0to1h"].Count++
			a.ExpireStats["exp0to1h"].Memory = a.ExpireStats["exp0to1h"].Memory + o.GetSize()
		case h <= 3 && h > 1:
			a.ExpireStats["exp1to3h"].Count++
			a.ExpireStats["exp1to3h"].Memory = a.ExpireStats["exp1to3h"].Memory + o.GetSize()
		case h <= 12 && h > 3:
			a.ExpireStats["exp3to12h"].Count++
			a.ExpireStats["exp3to12h"].Memory = a.ExpireStats["exp3to12h"].Memory + o.GetSize()
		case h <= 24 && h > 12:
			a.ExpireStats["exp12to24h"].Count++
			a.ExpireStats["exp12to24h"].Memory = a.ExpireStats["exp12to24h"].Memory + o.GetSize()
		case h <= 24*3 && h > 24:
			a.ExpireStats["exp1to2d"].Count++
			a.ExpireStats["exp1to2d"].Memory = a.ExpireStats["exp1to2d"].Memory + o.GetSize()
		case h <= 24*7 && h > 24*3:
			a.ExpireStats["exp3to7d"].Count++
			a.ExpireStats["exp3to7d"].Memory = a.ExpireStats["exp3to7d"].Memory + o.GetSize()
		case h > 24*7:
			a.ExpireStats["exp7dplus"].Count++
			a.ExpireStats["exp7dplus"].Memory = a.ExpireStats["exp7dplus"].Memory + o.GetSize()
		}
	}
}

func updatePrefixStats(a *RDBAnalysis, o parser.RedisObject) {
	for level := 1; level <= 3; level++ {
		prefix := GetPrefixMulti(o.GetKey(), level)
		if prefix == "" {
			continue
		}
		if _, ok := a.PrefixMemMap[prefix]; ok {
			a.PrefixMemMap[prefix].Elements = a.PrefixMemMap[prefix].Elements + o.GetElemCount()
			a.PrefixMemMap[prefix].Size = a.PrefixMemMap[prefix].Size + o.GetSize()
		} else {
			a.PrefixMemMap[prefix] = &KeyInfo{
				DB:       o.GetDBIndex(),
				Key:      prefix,
				Type:     o.GetType(),
				Size:     o.GetSize(),
				Elements: o.GetElemCount(), // 如果支持
			}
		}
	}
}

func analyseKey(analysis *RDBAnalysis, o parser.RedisObject) {
	updateBasicStats(analysis, o) // 总数、总内存、类型
	updateExpireStats(analysis, o)
	updatePrefixStats(analysis, o)
	analysis.TotalMemoryReable = formatBytes(analysis.TotalMemory)
}

// App struct
type App struct {
	ctx         context.Context
	rdbAnalysis map[string]*RDBAnalysis
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.rdbAnalysis = make(map[string]*RDBAnalysis)
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) ParseRDB(fileNames []string) string {
	fmt.Println(fileNames)
	for _, filepath := range fileNames {

		rdbFile, err := os.Open(filepath)
		if err != nil {
			panic("open dump.rdb failed")
		}
		defer func() {
			_ = rdbFile.Close()
		}()

		decoder := parser.NewDecoder(rdbFile)

		a.rdbAnalysis[filepath] = &RDBAnalysis{
			FileName:    filepath,
			TypeStats:   make(map[string]*TypeStat),
			ExpireStats: make(map[string]*TypeStat),
		}
		a.rdbAnalysis[filepath].TopKesHeap = NewTopNKeys(500)
		a.rdbAnalysis[filepath].PrefixTop500Heap = NewTopNKeys(500)
		a.rdbAnalysis[filepath].PrefixMemMap = make(map[string]*KeyInfo)
		for _, t := range []string{"string", "list", "set", "zset", "hash"} {
			a.rdbAnalysis[filepath].TypeStats[t] = &TypeStat{}
		}

		for _, t := range []string{"noexpire", "expired", "exp0to1h", "exp1to3h", "exp3to12h", "exp12to24h", "exp1to2d", "exp3to7d", "exp7dplus"} {
			a.rdbAnalysis[filepath].ExpireStats[t] = &TypeStat{}
		}

		err = decoder.Parse(func(o parser.RedisObject) bool {
			analyseKey(a.rdbAnalysis[filepath], o)

			return true
		})
		if err != nil {
			panic(err)
		}
		a.rdbAnalysis[filepath].TopKeys = a.rdbAnalysis[filepath].TopKesHeap.Items()
		for _, prefix := range a.rdbAnalysis[filepath].PrefixMemMap {
			prefix.ReadbleSize = formatBytes(prefix.Size)
			a.rdbAnalysis[filepath].PrefixTop500Heap.Add(prefix)
		}
		a.rdbAnalysis[filepath].TopPrefix500Keys = a.rdbAnalysis[filepath].PrefixTop500Heap.Items()
	}

	return "解析成功"
}

// 获取已保存的结果（可以多次调用，不会丢失）
func (a *App) GetParsedKeys(filename string) *RDBAnalysis {
	fmt.Println(filename)
	return a.rdbAnalysis[filename]
}

// 打开文件对话框
func (a *App) OpenFile() []string {
	file, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择一个文件",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "文本文件 (*.rdb)",
				Pattern:     "*.rdb",
			},
		},
	})
	if err != nil {
		fmt.Println("出错: %v", err)
	}

	return file
}
