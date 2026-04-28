import fs from "fs";
import path from "path";
import os from "os";

const CONFIG_DIR = path.join(os.homedir(), ".endgit");
const CONFIG_FILE = path.join(CONFIG_DIR, "config.json");

export interface EndGitConfig {
  apiToken?: string;
  apiUrl?: string;
}

export function getConfig(): EndGitConfig {
  if (!fs.existsSync(CONFIG_FILE)) {
    return { apiUrl: "http://localhost:4000" }; // Default API URL
  }
  try {
    const data = fs.readFileSync(CONFIG_FILE, "utf-8");
    return JSON.parse(data);
  } catch {
    return { apiUrl: "http://localhost:4000" };
  }
}

export function saveConfig(config: Partial<EndGitConfig>) {
  if (!fs.existsSync(CONFIG_DIR)) {
    fs.mkdirSync(CONFIG_DIR, { recursive: true });
  }
  const current = getConfig();
  const updated = { ...current, ...config };
  fs.writeFileSync(CONFIG_FILE, JSON.stringify(updated, null, 2), "utf-8");
}
