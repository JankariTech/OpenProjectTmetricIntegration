# Preparation
## As the admin
There are some steps that need to be done as the tmetric admin and the OpenProject admin.

### tmetric
- get the client ID by selecting the client, whose data you want to sync with OpenProject, from Manage->Clients and copying the last number from the URL. For example, if the URL is `https://app.tmetric.com/#/account/12345/clients/67890`, the client ID is `67890`.
- create a dummy project. You can give it any name you like. No time will be logged to this project, but it's needed to create links to OpenProject.
- get the ID of the dummy project. For example, if the URL is `https://app.tmetric.com/#/account/12345/projects/989898` the project ID is `989898`.
- create work types. The names of the Work Types **must match** the names of the time tracking activities in OpenProject.
- assign the work types to the project which data you want to sync to OpenProject.
- create a tag called `transferred-to-openproject`. This tag will be used to mark time entries that have been transferred to OpenProject.

### OpenProject
- enable the 'Time and costs' module for the projects you want to sync with tmetric

## As the user
### get tokens
#### OpenProject
- go to your profile (by clicking on your avatar in the top right corner)
- click on `My account`
- click on `Access tokens`
- click on `+ API token`
- give the token a name and click on `Create`
- copy the token

#### t-metric
- go to your profile (by clicking on your avatar in the top right corner)
- click on `Profile Settings`
- click on `Get new API token`
- copy the token

### configure

in your home folder create a file called `.OpenProjectTmetricIntegration.yaml` with the following content (enter your tokens):

```yaml
openproject:
  url: 'https://community.openproject.org'
  token: <openproject token>
tmetric:
  token: <t-metric token>
  clientId: <id of the client in tmetric>
  dummyProjectId: <id of the dummy project the admin created>
```

### run

#### check and fix the time entries in tmetric
```bash
go run main.go check tmetric
```

The tool with show you any invalid or inconsistent time entries and ask you for data to fix them.
This step will only adjust data on tmetric, no data will be written to OpenProject.

#### transfer the time entries to OpenProject
```bash
go run main.go copy
```

This will copy the time entries from tmetric to OpenProject. The time entries will be marked with the tag `transferred-to-openproject` in tmetric. Any entry that already has this tag will be skipped.

#### validate if the data in tmetric and OpenProject is consistent
```bash
go run main.go diff
```
This will show you a table with the time entries in both systems and if there is any difference in the logged time per day. Check if the data is correct.

1. If some data is in tmetric but not in OpenProject, run the `copy` command again.
2. If some data is in OpenProject but not in tmetric, delete or edit it in OpenProject. To do so use the [cost-report feature](https://www.openproject.org/docs/user-guide/time-and-costs/reporting/).
3. If you want to sync data again from tmetric to OpenProject, remove the `transferred-to-openproject` tag from the time entries in tmetric. **This will create new entries in OpenProject and by that might lead to duplication.**

#### work for a specific time period

By default, the script will work with the current calendar month, but the start and end date can be configured with the `--start` and `--end` flags. The date format is `YYYY-MM-DD`.

