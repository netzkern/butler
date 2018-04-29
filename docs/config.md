# Butler config

## The butler.yml file

```yml
templates:
  - name:                           The template name (string, required)
    url:                            The remote git or local file path to the template (string, required)

variables:
  test:                             The value for custom variable

confluence:
  templates:
    - name: software                The template name (string, required)
      pages:
        - name: Development         The page name (string, required)
          children:                 The children pages (page)
            - name: Architecture
            - name: Getting Started
```

## Custom variables

You can define custom variables to use them inside project templates. Custom template variables have priority over local variables.

## Config places

Butler searches for three different places for a `butler.yml` file.

* From your user space `~/.butler/butler.yml`
* From your current working directory `butler.yml`
* From the `BUTLER_CONFIG_URL` environment variable (Support also local paths)

The order above displays the merge order.
