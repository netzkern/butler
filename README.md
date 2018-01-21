<p align="center">
<img src="https://raw.githubusercontent.com/netzkern/butler/master/logo.png" alt="butler" style="max-width:100%;">
</p>

Welcome to Butler, your personal assistent to scaffolding your projects.
Shipped with binaries for Mac, Win and Linux (64-bit).

# Usage

1. [Download here](https://github.com/netzkern/butler/releases)
2. Install in `PATH`
3. Run `butler`

# Commands

* [Project Templates](#project-templates)
* [Auto Update](#auto-update)
* [Version](#version)

## Project Templates

This command will create a new project based on the selected template. It will replaces all variables in files and filenames. Project templates are listed in the `butler.yml` file. Butler is shipped with a default config.

### Available variables:

- `Project.Name`: Project name
- `Project.Description`: Project description
- `Date`: Current Date (RFC3339)
- `Year`: Current year

You can specify custom variables in the `butler.yml` file. They can be accessed by `butler{ .Vars.company }`.

## Auto Update

Butler is able to update itself. The latest Github release is used.

## Version

Print the version of Butler.

## Configuration

Stay up-to-date with new templates without to update your config manually just set the environment variable `BUTLER_CONFIG_URL` to [butler.yml on master](https://raw.githubusercontent.com/netzkern/butler/master/butler.yml) and both configs are merged.

## What Butler template looks like ?

[example-project-template](https://github.com/netzkern/example-project-template)

### Credits

<div>Icons made by <a href="http://www.freepik.com" title="Freepik">Freepik</a> from <a href="https://www.flaticon.com/" title="Flaticon">www.flaticon.com</a> is licensed by <a href="http://creativecommons.org/licenses/by/3.0/" title="Creative Commons BY 3.0" target="_blank">CC 3.0 BY</a></div>