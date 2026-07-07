import { existsSync, mkdirSync, readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import net from "node:net";
import { spawn } from "node:child_process";

export const rootDir = new URL("..", import.meta.url);
export const serverDir = new URL("../raglens-server/", import.meta.url);
const cacheDir = new URL("../.cache/", import.meta.url);

export function loadEnv(file = new URL("../.env", import.meta.url)) {
  if (!existsSync(file)) return;

  const lines = readFileSync(file, "utf8").split(/\r?\n/);
  for (const line of lines) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) continue;

    const separator = trimmed.indexOf("=");
    if (separator === -1) continue;

    const key = trimmed.slice(0, separator).trim();
    const value = stripQuotes(trimmed.slice(separator + 1).trim());
    if (key && process.env[key] === undefined) {
      process.env[key] = value;
    }
  }
}

export function databaseEndpoint() {
  const raw = process.env.RAGLENS_DATABASE_URL ?? "postgres://raglens:raglens@localhost:5432/raglens?sslmode=disable";
  const url = new URL(raw);
  return {
    url: raw,
    host: url.hostname || "localhost",
    port: Number(url.port || 5432),
  };
}

export async function run(command, args, options = {}) {
  if (command === "go") configureGoEnv();

  await new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      stdio: "inherit",
      shell: options.shell ?? process.platform === "win32",
      env: process.env,
      ...options,
    });

    child.on("error", reject);
    child.on("exit", (code) => {
      if (code === 0) {
        resolve();
      } else {
        reject(new Error(`${command} ${args.join(" ")} exited with code ${code}`));
      }
    });
  });
}

export function spawnLongRunning(command, args, options = {}) {
  if (command === "go") configureGoEnv();

  return spawn(command, args, {
    stdio: "inherit",
    shell: options.shell ?? process.platform === "win32",
    env: process.env,
    ...options,
  });
}

export function runNpm(args, options = {}) {
  const invocation = npmInvocation(args);
  return run(invocation.command, invocation.args, { ...options, shell: false });
}

export function spawnNpm(args, options = {}) {
  const invocation = npmInvocation(args);
  return spawnLongRunning(invocation.command, invocation.args, { ...options, shell: false });
}

export function configureGoEnv() {
  const goWork = fileURLToPath(new URL("go-work/", cacheDir));
  const goBuild = fileURLToPath(new URL("go-build/", cacheDir));
  const goMod = fileURLToPath(new URL("go-mod/", cacheDir));

  mkdirSync(goWork, { recursive: true });
  mkdirSync(goBuild, { recursive: true });
  mkdirSync(goMod, { recursive: true });

  process.env.GOPATH ??= goWork;
  process.env.GOCACHE ??= goBuild;
  process.env.GOMODCACHE ??= goMod;
}

export async function waitForTcp(host, port, timeoutMs = 30000) {
  const deadline = Date.now() + timeoutMs;

  while (Date.now() < deadline) {
    if (await canConnect(host, port)) return;
    await sleep(500);
  }

  throw new Error(`Timed out waiting for ${host}:${port}`);
}

function canConnect(host, port) {
  return new Promise((resolve) => {
    const socket = net.createConnection({ host, port });
    socket.setTimeout(1200);
    socket.once("connect", () => {
      socket.destroy();
      resolve(true);
    });
    socket.once("timeout", () => {
      socket.destroy();
      resolve(false);
    });
    socket.once("error", () => resolve(false));
  });
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function stripQuotes(value) {
  if ((value.startsWith('"') && value.endsWith('"')) || (value.startsWith("'") && value.endsWith("'"))) {
    return value.slice(1, -1);
  }
  return value;
}

function npmInvocation(args) {
  const npmCli = process.env.npm_execpath;
  if (npmCli) {
    return {
      command: process.execPath,
      args: [npmCli, ...args],
    };
  }

  return {
    command: "npm",
    args,
  };
}
