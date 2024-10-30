## get tokens
### OpenProject
- go to your profile (by clicking on your avatar in the top right corner)
- click on `My account`
- click on `Access tokens`
- click on `+ API token`
- give the token a name and click on `Create`
- copy the token

### t-metric
- go to your profile (by clicking on your avatar in the top right corner)
- click on `Profile Settings`
- click on `Get new API token`
- copy the token

## configure

in your home folder create a file called `.OpenProjectTmetricIntegration.yaml` with the following content (enter your tokens):

```yaml
openproject:
  url: 'https://community.openproject.org'
  token: <openproject token>
tmetric:
  token: <t-metric token>
  clientId: 124390
  dummyProjectId: 948448
```

## run

```bash
go run main.go check tmetric
```

by default the script will check the current calendar month, but the start and end date can be configured with the `--start` and `--end` flags. The date format is `YYYY-MM-DD`.

