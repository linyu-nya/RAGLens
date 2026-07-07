import { rootDir, run, runNpm } from "./lib.mjs";

await run("go", ["build", "./cmd/raglens-server"], { cwd: new URL("../raglens-server/", import.meta.url) });
await runNpm(["run", "build", "--workspace", "raglens-web"], { cwd: rootDir });
