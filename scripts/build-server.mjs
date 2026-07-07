import { run, serverDir } from "./lib.mjs";

await run("go", ["build", "./cmd/raglens-server"], { cwd: serverDir });
