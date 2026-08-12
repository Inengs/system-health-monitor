export namespace main {
	
	export class Alert {
	    pid: number;
	    name: string;
	    reason: string;
	    time: string;
	
	    static createFrom(source: any = {}) {
	        return new Alert(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pid = source["pid"];
	        this.name = source["name"];
	        this.reason = source["reason"];
	        this.time = source["time"];
	    }
	}
	export class ProcInfo {
	    pid: number;
	    name: string;
	    friendly: string;
	    cpuPercent: number;
	    memPercent: number;
	    safeToClose: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProcInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pid = source["pid"];
	        this.name = source["name"];
	        this.friendly = source["friendly"];
	        this.cpuPercent = source["cpuPercent"];
	        this.memPercent = source["memPercent"];
	        this.safeToClose = source["safeToClose"];
	    }
	}
	export class Snapshot {
	    topByCpu: ProcInfo[];
	    topByMem: ProcInfo[];
	    alerts: Alert[];
	
	    static createFrom(source: any = {}) {
	        return new Snapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.topByCpu = this.convertValues(source["topByCpu"], ProcInfo);
	        this.topByMem = this.convertValues(source["topByMem"], ProcInfo);
	        this.alerts = this.convertValues(source["alerts"], Alert);
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

