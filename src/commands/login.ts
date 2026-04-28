import pc from "picocolors";
import prompts from "prompts";
import { saveConfig } from "../utils/config";

export async function loginCommand() {
  console.log(pc.bold(pc.cyan("Authenticate with EndGit\n")));

  const response = await prompts([
    {
      type: "password",
      name: "token",
      message: "Enter your Personal Access Token (PAT):",
      validate: (value) => (value.length > 10 ? true : "Invalid token format"),
    },
  ]);

  if (!response.token) {
    console.log(pc.red("Login cancelled."));
    process.exit(1);
  }

  saveConfig({ apiToken: response.token });
  console.log(pc.green("✔ Successfully logged in. Token saved locally."));
}
