import { rootDir, runNpm } from "./lib.mjs";

await runNpm(["run", "build", "--workspace", "raglens-web"], { cwd: rootDir });
