<p align="center">
<img src="/logo.png" alt="butler" style="max-width:100%;">
</p>

Butler is an automatation tool to scaffold new projects in only a few seconds.
We provide a powerful interactive cli. When you create a project template you
can create a [`Survey`](/docs/templateSurveys.md). Surveys are used to collect
informations from the users to generate individual templates. Beside
templating we also plan to integrate common commands for popular Project Management
Tools like Jira, Confluence in Butler.

> Bootstraping projects should be fun!

## Features
- ✔︎ Template Surveys
- ✔︎ Conditional files and folders
- ✔︎ After hooks for post-processing
- :sparkles: **Maintanance:** Auto Update, Distributed configs
- :star2: **Confluence:** Create spaces with preconfigured page tree

## Principles
- Project Templates are simple git repositories
- Everything is a template you don't have to deal with `/template` directories
or `.tmpl` files
- Required informations are asked during the bootstrapping process

## Usage

1. [Download here](https://github.com/netzkern/butler/releases)
2. Install in `PATH`
3. Run `butler`

## Documentation

* [**Config**](/docs/config.md)
* [**Template Surveys**](/docs/templateSurveys.md)
* [**Template Syntax**](/docs/templateSyntax.md)
* [**Git Hooks**](/docs/gitHooks.md)
* [**Confluence**](/docs/confluence.md)
* [**Debugging**](/docs/debugging.md)
* [**Commands**](#commands)

## Commands

- **Create Project:** This command will create a new project based on the selected template.
- **Create Git Hooks:** This command will install all selected hooks.
- **Create Confluence Space:** This command will create a public or private confluence space based on the selected template.
- **Maintanance:**
  - **Dump config:** Prints the final butler config in the terminal.
  - **Auto Update:** This command will update Butler to the latest version.
  - **Report a bug:** This command will open a new Github issue.
  - **Version:** This command will return the current version of Butler.

## Maintenance across teams

- Butler is able to update itself. The latest Github release is used.
- Stay up-to-date with new templates without to update your config manually just set the environment variable `BUTLER_CONFIG_URL` to [butler.yml on master](https://raw.githubusercontent.com/netzkern/butler/master/butler.yml) and both configs are merged.

## What Butler template looks like ?

[example-project-template](https://github.com/netzkern/example-project-template)

## Lead Maintainers

- [**Dustin Deus**](https://github.com/StarpTech), <https://twitter.com/dustindeus>, <https://www.npmjs.com/~starptech>

## Acknowledgements

This project is kindly sponsored by [netzkern](http://netzkern.de). We're [hiring!](http://karriere.netzkern.de/)

## License

Licensed under [MIT](./LICENSE).

### Credits

<div>Icons made by <a href="http://www.freepik.com" title="Freepik">Freepik</a> from <a href="https://www.flaticon.com/" title="Flaticon">www.flaticon.com</a> is licensed by <a href="http://creativecommons.org/licenses/by/3.0/" title="Creative Commons BY 3.0" target="_blank">CC 3.0 BY</a></div>
