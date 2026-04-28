import fs from "fs";
import path from "path";
import pc from "picocolors";
import ora from "ora";
import archiver from "archiver";
import FormData from "form-data";
import { getApiClient } from "../utils/api";

export async function publishCommand() {
  console.log(pc.bold(pc.cyan("Publishing to EndGit...\n")));
  
  const pluginJsonPath = path.join(process.cwd(), "plugin.json");
  const endgitJsonPath = path.join(process.cwd(), "endgit.json");

  if (!fs.existsSync(pluginJsonPath)) {
    console.error(pc.red("Error: plugin.json not found in current directory."));
    console.log(`Run ${pc.white("endgit init")} to create one.`);
    process.exit(1);
  }

  const spinner = ora("Reading project configuration...").start();

  try {
    const pluginJson = JSON.parse(fs.readFileSync(pluginJsonPath, "utf-8"));
    let endgitJson: any = {};
    if (fs.existsSync(endgitJsonPath)) {
      endgitJson = JSON.parse(fs.readFileSync(endgitJsonPath, "utf-8"));
    }

    spinner.text = `Preparing to publish ${pluginJson.name} v${pluginJson.version}...`;
    
    const api = getApiClient();

    // Check if plugin exists, if not create draft
    try {
      await api.get(`/api/v1/plugins/${pluginJson.name}`);
    } catch {
      await api.post(`/api/v1/plugins`, {
        name: pluginJson.name,
        displayName: endgitJson.displayName || pluginJson.name,
        description: pluginJson.description || "A new Endstone plugin",
        pluginType: endgitJson.pluginType || "PYTHON"
      });
    }

    spinner.text = "Zipping source code...";
    
    // Create ZIP in memory or temp file
    const zipPath = path.join(process.cwd(), ".endgit_temp.zip");
    const output = fs.createWriteStream(zipPath);
    const archive = archiver('zip', { zlib: { level: 9 } });

    await new Promise<void>((resolve, reject) => {
      output.on('close', resolve);
      archive.on('error', reject);
      archive.pipe(output);

      // Add all files except ignored ones
      archive.glob('**/*', {
        cwd: process.cwd(),
        ignore: ['node_modules/**', 'venv/**', '.git/**', '.endgit_temp.zip', '__pycache__/**', '*.pyc']
      });

      archive.finalize();
    });

    spinner.text = "Uploading source to EndGit...";

    const formData = new FormData();
    formData.append("source", fs.createReadStream(zipPath));
    // Try to get local git commit if available
    try {
      const { execSync } = require("child_process");
      const commitHash = execSync("git rev-parse HEAD").toString().trim();
      const branch = execSync("git rev-parse --abbrev-ref HEAD").toString().trim();
      formData.append("commitHash", commitHash);
      formData.append("branch", branch);
    } catch (e) {
      // Ignore if not a git repo
    }

    await api.post(`/api/v1/plugins/${pluginJson.name}/publish`, formData, {
      headers: { ...formData.getHeaders() }
    });

    // Cleanup temp zip
    fs.unlinkSync(zipPath);

    spinner.succeed(`Successfully uploaded ${pc.bold(pluginJson.name)} source to EndGit!`);
    console.log(pc.gray(`\nThe build is now running in the CI/CD pipeline.`));
    console.log(pc.gray(`View live build logs at: `) + pc.underline(pc.cyan(`http://localhost:3000/builds`)));

  } catch (error: any) {
    spinner.fail("Publish failed.");
    console.error(pc.red(error.response?.data?.error || error.message));
    if (error.response?.status === 401) {
      console.log(`Run ${pc.white("endgit login")} to authenticate.`);
    }
    // Cleanup if exists
    const zipPath = path.join(process.cwd(), ".endgit_temp.zip");
    if (fs.existsSync(zipPath)) fs.unlinkSync(zipPath);
  }
}
