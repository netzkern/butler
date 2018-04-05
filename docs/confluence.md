# Butler confluence

Create a space or a complete page tree based on templates.

## Configuration

Before you can use confluence commands you have to setup authentication & authorization.

```
BUTLER_CONFLUENCE_URL=https://confluence.company.de     The base url of your confluence server (string, required)
BUTLER_CONFLUENCE_AUTH_METHOD=basic                     The authentication method (string, required)
BUTLER_CONFLUENCE_BASIC_AUTH=username,password          The basic authentication credentials comma seperated (string, required)
```

### Confluence permission

You have to assign each dev the global permission "Create Space(s)".
When a space is created, the creator automatically has the `Admin` permission for that space and can perform space-wide administrative functions.

## Commands

* Create Confluence Space

  This command will create a new space. You can specify the space name and description.
  The Space Key is generated from the Space name and converted to `camelCase` e.g `my project` to Key `myProject`. In the final stage you can select your template to generate the space page-tree.

  **Questions:**

  * What's the name of the space?
  * What's the description of the space?
  * Do you want to create a public space?
  * Please select a template (If you want to skip this select `None`)

## Create page tree based on template

You can configure the tree structure of your space with the help of templates. These templates can be configured in the `butler.yml` file and has the following structure.

```yml
confluence:
  templates:
    - name: None
    - name: Software Project
      pages:
        - name: Introduction
        - name: Product Requirements
        - name: Meeting notes
        - name: Retroperspectives
        - name: Shared links
        - name: Development
        - name: Infrastruktur
        - name: Design
          children:
            - name: test1
            - name: test2
              children:
                - name: test3
                - name: test4
```
