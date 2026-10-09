// Reachable syntactic completions. Jump targets are Adamic parser indices.
export class Completion {
    normal: boolean;
    thrown = false;
    readonly breaks: number[] = [];
    readonly continues: number[] = [];
    constructor(normal: boolean) {
        this.normal = normal;
    }
    merge(other: Completion): void {
        if(other.normal) {
            this.normal = true;
        }
        if(other.thrown) {
            this.thrown = true;
        }
        for(const target of other.breaks) {
            if(!this.breaks.includes(target)) {
                this.breaks.push(target);
            }
        }
        for(const target of other.continues) {
            if(!this.continues.includes(target)) {
                this.continues.push(target);
            }
        }
    }
    outward(other: Completion, target: number): void {
        if(other.thrown) {
            this.thrown = true;
        }
        for(const jump of other.breaks) {
            if(jump !== target && !this.breaks.includes(jump)) {
                this.breaks.push(jump);
            }
        }
        for(const jump of other.continues) {
            if(jump !== target && !this.continues.includes(jump)) {
                this.continues.push(jump);
            }
        }
    }
}
