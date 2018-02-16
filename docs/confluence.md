# Butler confluence
Before you can use confluence commands you have to setup authentication & authorization.

## Configuration

```
BUTLER_CONFLUENCE_URL=https://confluence.company.de     The base url of your confluence server (string, required)
BUTLER_CONFLUENCE_AUTH_METHOD=basic                     The authentication method (string, required)
BUTLER_CONFLUENCE_BASIC_AUTH=username,password          The basic authentication credentials comma seperated (string, required)
```

### Confluence permission
You have to assign each dev the global permission "Create Space(s)".
When a space is created, the creator automatically has the 'Admin' permission for that space and can perform space-wide administrative functions.

## Commands

- Create Confluence Space

  This command will create a new space. You can specify the space name and description.
  The Space Key is generated from the Space name and converted to `chainCase` e.g `my project` to Key `my-project`

  **Questions:**
  - What's the name of the space?
  - What's the description of the space?
  - Do you want to create a private space?