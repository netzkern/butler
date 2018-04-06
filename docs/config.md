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

#### Custom variables

You can define custom variables to use them inside project templates. Custom template variables have priority over local variables.

### Distribute config

You can set the environment variable `BUTLER_CONFIG_URL` to any url to load your config from an external or local storage.
This make it easy to distribute template updates across a company. You local configuration is merged.
