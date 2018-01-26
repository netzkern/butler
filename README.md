<p align="center">
<img src="https://raw.githubusercontent.com/netzkern/butler/master/logo.png" alt="butler" style="max-width:100%;">
</p>

Butler is an automatation tool to scaffold new projects in only a few seconds. We provide a powerful interactive cli.
When you create a project template you can create a [`Survey`](/docs/templateSurveys.md). Surveys are used to collect informations from the users to generate individual templates. Beside templating we also plan to integrate common tasks of popular project managents tools like Jira, Confuence in Butler.

## Usage

1. [Download here](https://github.com/netzkern/butler/releases)
2. Install in `PATH`
3. Run `butler`

## Documentation

* <a href="/docs/templateSurveys.md"><code><b>Template Surveys</b></code></a>
* <a href="/docs/templateSyntax.md"><code><b>Template Syntax</b></code></a>
* <a href="/docs/gitHooks.md"><code><b>Git Hooks</b></code></a>
* <a href="/docs/debugging.md"><code><b>Debugging</b></code></a>
* <a href="#commands"><code><b>Commands</b></code></a>

## Commands

- **Project Templates:** This command will create a new project based on the selected template.
- **Install Git Hooks:** This command will install all selected hooks.
- **Auto Update:** This command will update Butler to the latest version.
- **Version:** This command will return the current version of Butler.

## Maintenance

- Butler is able to update itself. The latest Github release is used.
- Stay up-to-date with new templates without to update your config manually just set the environment variable `BUTLER_CONFIG_URL` to [butler.yml on master](https://raw.githubusercontent.com/netzkern/butler/master/butler.yml) and both configs are merged.

## What Butler template looks like ?

[example-project-template](https://github.com/netzkern/example-project-template)

## Roadmap

- [ ] Create templates for Sitecore, Kentico or other components
- [ ] Jira Integration
- [ ] Confluence Integration

### Credits

<div>Icons made by <a href="http://www.freepik.com" title="Freepik">Freepik</a> from <a href="https://www.flaticon.com/" title="Flaticon">www.flaticon.com</a> is licensed by <a href="http://creativecommons.org/licenses/by/3.0/" title="Creative Commons BY 3.0" target="_blank">CC 3.0 BY</a></div>
