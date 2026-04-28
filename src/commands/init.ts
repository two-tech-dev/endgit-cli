import fs from "fs";
import path from "path";
import pc from "picocolors";
import prompts from "prompts";

export async function initCommand() {
  console.log(pc.bold(pc.cyan("Initialize Endstone Plugin\n")));

  const currentDir = path.basename(process.cwd());

  const response = await prompts([
    {
      type: "text",
      name: "name",
      message: "Plugin Name (slug):",
      initial: currentDir.toLowerCase().replace(/[^a-z0-9-]/g, "-"),
    },
    {
      type: "text",
      name: "displayName",
      message: "Display Name:",
      initial: currentDir,
    },
    {
      type: "text",
      name: "version",
      message: "Version:",
      initial: "1.0.0",
    },
    {
      type: "text",
      name: "description",
      message: "Description:",
    },
    {
      type: "text",
      name: "api",
      message: "Required Endstone API version:",
      initial: "^0.5.0",
    },
    {
      type: "select",
      name: "type",
      message: "Plugin Type:",
      choices: [
        { title: "Python", value: "PYTHON" },
        { title: "C++", value: "CPP" },
      ],
    },
  ]);

  if (!response.name) {
    console.log(pc.red("Initialization cancelled."));
    process.exit(1);
  }

  const pluginJson = {
    name: response.name,
    version: response.version,
    description: response.description,
    api: [response.api],
    main: response.type === "PYTHON" ? "src.main" : "EndstonePlugin",
  };

  const filePath = path.join(process.cwd(), "plugin.json");
  fs.writeFileSync(filePath, JSON.stringify(pluginJson, null, 2), "utf-8");

  // Create endgit metadata file
  const endgitJson = {
    displayName: response.displayName,
    pluginType: response.type,
  };
  fs.writeFileSync(path.join(process.cwd(), "endgit.json"), JSON.stringify(endgitJson, null, 2), "utf-8");

  console.log(pc.green("\n✔ Successfully created plugin.json and endgit.json."));
  console.log(pc.gray(`You can now run ${pc.white("endgit publish")} to push your plugin.`));
}
