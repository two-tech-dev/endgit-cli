import fs from "fs";
import path from "path";
import pc from "picocolors";
import ora from "ora";
import { getApiClient } from "../utils/api";

export async function installCommand(input: string) {
  // Parse plugin@commit_hash syntax
  let pluginName = input;
  let commitHash: string | null = null;
  
  if (input.includes("@")) {
    const parts = input.split("@");
    pluginName = parts[0];
    commitHash = parts[1];
  }

  const spinner = ora(
    commitHash
      ? `Fetching dev build for "${pluginName}" at commit ${commitHash.slice(0, 7)}...`
      : `Fetching info for plugin "${pluginName}"...`
  ).start();

  try {
    const api = getApiClient();
    
    if (commitHash) {
      // ── Dev Build Install (by commit hash) ──
      spinner.text = `Searching builds for commit ${pc.yellow(commitHash.slice(0, 7))}...`;

      // Get builds for this plugin, filtered by commit
      const buildsRes = await api.get(`/api/v1/builds/plugin/${pluginName}`);
      const builds = buildsRes.data.data?.builds || [];
      
      const targetBuild = builds.find((b: any) => 
        b.commitHash && b.commitHash.startsWith(commitHash!) && b.status === "SUCCESS"
      );

      if (!targetBuild) {
        spinner.fail(`No successful build found for commit ${commitHash.slice(0, 7)}`);
        console.log(pc.yellow("⚠️  Tip: Use 'endgit search " + pluginName + "' to find available builds."));
        return;
      }

      console.log(pc.yellow(`\n⚠️  WARNING: This is a DEV BUILD (not reviewed). Use at your own risk.`));
      spinner.text = `Downloading dev build #${targetBuild.buildNumber}...`;

      let downloadUrl = targetBuild.artifactUrl;
      if (process.platform === "win32" && targetBuild.artifactUrlWin) {
        downloadUrl = targetBuild.artifactUrlWin;
      } else if ((process.platform === "linux" || process.platform === "darwin") && targetBuild.artifactUrlLinux) {
        downloadUrl = targetBuild.artifactUrlLinux;
      }

      if (!downloadUrl) {
        spinner.fail(`Build #${targetBuild.buildNumber} does not have a valid artifact for your operating system.`);
        return;
      }

      // Setup plugins directory
      const pluginsDir = path.join(process.cwd(), "plugins");
      if (!fs.existsSync(pluginsDir)) fs.mkdirSync(pluginsDir);

      try {
        const downloadRes = await api.get(downloadUrl, {
          responseType: "arraybuffer"
        });

        const contentDisposition = downloadRes.headers['content-disposition'];
        let fileName = `${pluginName}-build${targetBuild.buildNumber}-${commitHash.slice(0, 7)}.whl`;
        if (contentDisposition && contentDisposition.includes('filename=')) {
          fileName = contentDisposition.split('filename=')[1].replace(/"/g, '');
        }

        const filePath = path.join(pluginsDir, fileName);
        fs.writeFileSync(filePath, downloadRes.data);

        spinner.succeed(`Installed dev build ${pc.bold(pluginName)} #${targetBuild.buildNumber} (${commitHash.slice(0, 7)})`);
        console.log(pc.gray(`Saved to: ./plugins/${fileName}`));
        console.log(pc.yellow("⚠️  UNSTABLE — This build has NOT been reviewed."));
      } catch (dlError: any) {
        spinner.fail(`Failed to download dev build artifact: ${dlError.message}`);
      }

    } else {
      // ── Release Install (latest stable) ──
      const pluginRes = await api.get(`/api/v1/plugins/${pluginName}`);
      const plugin = pluginRes.data.data;
      
      if (!plugin.latestVersion) {
        spinner.fail(`Plugin "${pluginName}" has no published versions yet.`);
        return;
      }

      spinner.text = `Downloading ${pluginName} v${plugin.latestVersion}...`;

      const pluginsDir = path.join(process.cwd(), "plugins");
      if (!fs.existsSync(pluginsDir)) fs.mkdirSync(pluginsDir);

      try {
        const platform = process.platform === "win32" ? "windows" : "linux";
        const downloadRes = await api.get(`/api/v1/download/${pluginName}/${plugin.latestVersion}?platform=${platform}`, {
          responseType: "arraybuffer"
        });

        const contentDisposition = downloadRes.headers['content-disposition'];
        let fileName = `${pluginName}-${plugin.latestVersion}`;
        if (contentDisposition && contentDisposition.includes('filename=')) {
          fileName = contentDisposition.split('filename=')[1].replace(/"/g, '');
        } else {
          if (plugin.pluginType === "PYTHON") fileName += ".whl";
          else if (platform === "windows") fileName += ".dll";
          else fileName += ".so";
        }

        const filePath = path.join(pluginsDir, fileName);
        fs.writeFileSync(filePath, downloadRes.data);

        spinner.succeed(`Successfully installed ${pc.bold(pluginName)} v${plugin.latestVersion}`);
        console.log(pc.gray(`Saved to: ./plugins/${fileName}`));
      } catch {
        spinner.fail(`Artifact for ${pluginName} v${plugin.latestVersion} not found in storage.`);
      }
    }

  } catch (error: any) {
    spinner.fail("Installation failed.");
    if (error.response?.status === 404) {
      console.error(pc.red(`Error: Plugin "${pluginName}" not found.`));
    } else {
      console.error(pc.red(error.message));
    }
  }
}
