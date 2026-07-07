import { databaseEndpoint, loadEnv, run, serverDir } from "./lib.mjs";

loadEnv();

const database = databaseEndpoint();

await run(
  "go",
  [
    "run",
    "github.com/pressly/goose/v3/cmd/goose@latest",
    "-dir",
    "migrations",
    "postgres",
    database.url,
    "up",
  ],
  { cwd: serverDir },
);
