export namespace main {
	
	export class KeyInfo {
	    db: number;
	    key: string;
	    type: string;
	    size: number;
	    readble_size: string;
	    elements: number;
	    expire: string;
	
	    static createFrom(source: any = {}) {
	        return new KeyInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.db = source["db"];
	        this.key = source["key"];
	        this.type = source["type"];
	        this.size = source["size"];
	        this.readble_size = source["readble_size"];
	        this.elements = source["elements"];
	        this.expire = source["expire"];
	    }
	}
	export class TopNKeys {
	
	
	    static createFrom(source: any = {}) {
	        return new TopNKeys(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class TypeStat {
	    count: number;
	    memory: number;
	
	    static createFrom(source: any = {}) {
	        return new TypeStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.count = source["count"];
	        this.memory = source["memory"];
	    }
	}
	export class RDBAnalysis {
	    file_name: string;
	    TotalMemory: number;
	    total_memory: string;
	    total_keys: number;
	    type_stats: Record<string, TypeStat>;
	    expire_stats: Record<string, TypeStat>;
	    top_keys: KeyInfo[];
	    // Go type: TopNKeys
	    TopKesHeap?: any;
	    top_prefix_keys: KeyInfo[];
	    PrefixMemMap: Record<string, KeyInfo>;
	    // Go type: TopNKeys
	    PrefixTop500Heap?: any;
	
	    static createFrom(source: any = {}) {
	        return new RDBAnalysis(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.file_name = source["file_name"];
	        this.TotalMemory = source["TotalMemory"];
	        this.total_memory = source["total_memory"];
	        this.total_keys = source["total_keys"];
	        this.type_stats = this.convertValues(source["type_stats"], TypeStat, true);
	        this.expire_stats = this.convertValues(source["expire_stats"], TypeStat, true);
	        this.top_keys = this.convertValues(source["top_keys"], KeyInfo);
	        this.TopKesHeap = this.convertValues(source["TopKesHeap"], null);
	        this.top_prefix_keys = this.convertValues(source["top_prefix_keys"], KeyInfo);
	        this.PrefixMemMap = this.convertValues(source["PrefixMemMap"], KeyInfo, true);
	        this.PrefixTop500Heap = this.convertValues(source["PrefixTop500Heap"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

