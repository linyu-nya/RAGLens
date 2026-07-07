import { databaseEndpoint, loadEnv, rootDir, run, serverDir, spawnLongRunning, spawnNpm, waitForTcp } from "./lib.mjs";

loadEnv();

const database = databaseEndpoint();

await run("docker", ["compose", "up", "-d", "postgres"], { cwd: rootDir });
await waitForTcp(database.host, database.port);
await run("node", ["scripts/migrate.mjs"], { cwd: rootDir });

const backend = spawnLongRunning("go", ["run", "./cmd/raglens-server"], { cwd: serverDir });
const frontend = spawnNpm(["run", "dev", "--workspace", "raglens-web"], { cwd: rootDir });
const children = [backend, frontend];

for (const signal of ["SIGINT", "SIGTERM"]) {
  process.on(signal, () => {
    for (const child of children) child.kill(signal);
  });
}

await new Promise((resolve) => {
  for (const child of children) {
    child.on("exit", (code) => {
      for (const other of children) {
        if (other !== child && other.exitCode === null) other.kill();
      }
      process.exitCode = code ?? 0;
      resolve();
    });
  }
});
