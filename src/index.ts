#!/usr/bin/env node

import { Command } from "commander";
import pc from "picocolors";
import { loginCommand } from "./commands/login";
import { initCommand } from "./commands/init";
import { searchCommand } from "./commands/search";
import { installCommand } from "./commands/install";

const program = new Command();

program
  .name("endgit")
  .description("EndGit CLI - The package manager for Endstone plugins")
  .version("0.1.0");

program
  .command("login")
  .description("Authenticate with the EndGit platform")
  .action(loginCommand);

program
  .command("init")
  .description("Initialize a new Endstone plugin configuration")
  .action(initCommand);

program
  .command("search <query>")
  .description("Search for plugins in the EndGit registry")
  .action(searchCommand);

program
  .command("install <plugin>")
  .description("Download and install a plugin to the current directory")
  .action(installCommand);



// Fallback for unknown commands
program.on('command:*', function () {
  console.error(pc.red('Invalid command: %s\nSee --help for a list of available commands.'), program.args.join(' '));
  process.exit(1);
});

program.parse(process.argv);
