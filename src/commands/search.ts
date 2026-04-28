import pc from "picocolors";
import ora from "ora";
import { getApiClient } from "../utils/api";

export async function searchCommand(query: string) {
  const spinner = ora(`Searching for "${query}"...`).start();

  try {
    const api = getApiClient();
    const response = await api.get(`/api/v1/plugins?q=${encodeURIComponent(query)}`);
    const plugins = response.data.data;

    spinner.stop();

    if (plugins.length === 0) {
      console.log(pc.yellow(`No plugins found matching "${query}".`));
      return;
    }

    console.log(pc.bold(pc.cyan(`\nFound ${plugins.length} plugin(s):`)));
    
    // Simple tabular display
    console.log(pc.gray("-".repeat(80)));
    plugins.forEach((p: any) => {
      const typeStr = p.pluginType === "PYTHON" ? pc.green("[Py]") : pc.blue("[C++]");
      const nameStr = pc.bold(p.name.padEnd(20));
      const downloadsStr = pc.yellow(`${p.downloads} ⬇`);
      const versionStr = pc.magenta(`v${p.latestVersion || "?.?.?"}`);
      
      console.log(`${typeStr} ${nameStr} | ${versionStr} | ${downloadsStr}`);
      console.log(`     ${pc.gray(p.description)}`);
    });
    console.log(pc.gray("-".repeat(80)));
    console.log(`Run ${pc.white(`endgit install <plugin-name>`)} to install.\n`);

  } catch (error: any) {
    spinner.fail("Search failed.");
    console.error(pc.red(error.message));
  }
}
