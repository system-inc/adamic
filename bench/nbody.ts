// nbody: the Computer Language Benchmarks Game's planets, advanced in small steps. Floating-point
// arithmetic on the fields of a few objects: what a native compiler should win outright.

class Body {
    x: number;
    y: number;
    z: number;
    vx: number;
    vy: number;
    vz: number;
    readonly mass: number;

    constructor(x: number, y: number, z: number, vx: number, vy: number, vz: number, mass: number) {
        this.x = x;
        this.y = y;
        this.z = z;
        this.vx = vx;
        this.vy = vy;
        this.vz = vz;
        this.mass = mass;
    }
}

const solarMass = 4 * Math.PI * Math.PI;
const daysPerYear = 365.24;

function planets(): Body[] {
    return [
        new Body(0, 0, 0, 0, 0, 0, solarMass),
        new Body(
            4.8414314424647209,
            -1.16032004402742839,
            -1.03622044471123109e-1,
            1.66007664274403694e-3 * daysPerYear,
            7.69901118419740425e-3 * daysPerYear,
            -6.90460016972063023e-5 * daysPerYear,
            9.54791938424326609e-4 * solarMass,
        ),
        new Body(
            8.34336671824457987,
            4.12479856412430479,
            -4.03523417114321381e-1,
            -2.76742510726862411e-3 * daysPerYear,
            4.99852801234917238e-3 * daysPerYear,
            2.30417297573763929e-5 * daysPerYear,
            2.85885980666130812e-4 * solarMass,
        ),
        new Body(
            1.2894369562139131e1,
            -1.51111514016986312e1,
            -2.23307578892655734e-1,
            2.96460137564761618e-3 * daysPerYear,
            2.3784717395948095e-3 * daysPerYear,
            -2.96589568540237556e-5 * daysPerYear,
            4.36624404335156298e-5 * solarMass,
        ),
        new Body(
            1.53796971148509165e1,
            -2.59193146099879641e1,
            1.79258772950371181e-1,
            2.68067772490389322e-3 * daysPerYear,
            1.62824170038242295e-3 * daysPerYear,
            -9.5159225451971587e-5 * daysPerYear,
            5.15138902046611451e-5 * solarMass,
        ),
    ];
}

function offsetMomentum(bodies: Body[]): void {
    let px = 0;
    let py = 0;
    let pz = 0;
    for(const body of bodies) {
        px += body.vx * body.mass;
        py += body.vy * body.mass;
        pz += body.vz * body.mass;
    }
    const sun = bodies[0] ?? new Body(0, 0, 0, 0, 0, 0, solarMass);
    sun.vx = -px / solarMass;
    sun.vy = -py / solarMass;
    sun.vz = -pz / solarMass;
}

function advance(bodies: Body[], step: number): void {
    const count = bodies.length;
    for(let i = 0; i < count; i += 1) {
        const body = bodies[i] ?? new Body(0, 0, 0, 0, 0, 0, 0);
        for(let j = i + 1; j < count; j += 1) {
            const other = bodies[j] ?? new Body(0, 0, 0, 0, 0, 0, 0);
            const dx = body.x - other.x;
            const dy = body.y - other.y;
            const dz = body.z - other.z;
            const squared = dx * dx + dy * dy + dz * dz;
            const magnitude = step / (squared * Math.sqrt(squared));
            body.vx -= dx * other.mass * magnitude;
            body.vy -= dy * other.mass * magnitude;
            body.vz -= dz * other.mass * magnitude;
            other.vx += dx * body.mass * magnitude;
            other.vy += dy * body.mass * magnitude;
            other.vz += dz * body.mass * magnitude;
        }
    }
    for(const body of bodies) {
        body.x += step * body.vx;
        body.y += step * body.vy;
        body.z += step * body.vz;
    }
}

function energy(bodies: Body[]): number {
    let total = 0;
    const count = bodies.length;
    for(let i = 0; i < count; i += 1) {
        const body = bodies[i] ?? new Body(0, 0, 0, 0, 0, 0, 0);
        total += 0.5 * body.mass * (body.vx * body.vx + body.vy * body.vy + body.vz * body.vz);
        for(let j = i + 1; j < count; j += 1) {
            const other = bodies[j] ?? new Body(0, 0, 0, 0, 0, 0, 0);
            const dx = body.x - other.x;
            const dy = body.y - other.y;
            const dz = body.z - other.z;
            total -= (body.mass * other.mass) / Math.sqrt(dx * dx + dy * dy + dz * dz);
        }
    }
    return total;
}

const bodies = planets();
offsetMomentum(bodies);
console.log(energy(bodies).toFixed(9));
for(let step = 0; step < 1000000; step += 1) {
    advance(bodies, 0.01);
}
console.log(energy(bodies).toFixed(9));
