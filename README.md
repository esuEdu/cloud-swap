# cloud-swap

`cloud-swap` is a lightweight CLI to manage and switch between multiple AWS (and future multicloud) credentials quickly.  
Ideal for developers who use different AWS accounts for work, personal projects, study, clients, or temporary STS sessions.

## Features

-   Fast credential switching
-   Store multiple named profiles
-   Permanent and temporary (STS) credentials
-   Output environment variables for `eval $(...)` usage
-   Optional fuzzy selector (fzf)
-   Simple JSON storage model
-   Ready for multi-cloud in the future

## Credential Model (“Add Credit”)

Credentials are stored in:

```
~/.cloud-swap/credentials.json
```

### Schema example:

```json
{
	"name": "work",
	"provider": "aws",
	"access_key": "AWS_ACCESS_KEY_ID",
	"secret_key": "AWS_SECRET_ACCESS_KEY",
	"session_token": null,
	"region": "us-east-1",
	"output": "json"
}
```

### Temporary (STS) credential example:

```json
{
  "name": "temp-admin",
  "provider": "aws",
  "access_key": "ASIAxxxxxxxx",
  "secret_key": "xxxxxxxxxxxx",
  "session_token": "IQoJb3JpZ2luX2VjEP///////////",
  "expires_at": "2025-01-27T10:30:00Z",
  "region": "us-west-2"
  "output": "json"
}
```

## Full Example of `credentials.json`

```json
{
	"profiles": [
		{
			"name": "work",
			"provider": "aws",
			"access_key": "AKIAWORK123",
			"secret_key": "WORKSECRETXYZ",
			"region": "us-east-1"
            "output": "json"
		},
		{
			"name": "personal",
			"provider": "aws",
			"access_key": "AKIAPERSONAL123",
			"secret_key": "PERSONALSECRETXYZ",
			"region": "sa-east-1"
            "output": "json"
		},
		{
			"name": "study",
			"provider": "aws",
			"access_key": "AKIASTUDY123",
			"secret_key": "STUDYSECRETXYZ",
			"region": "us-west-2"
            "output": "json"
		},
		{
			"name": "temp-admin",
			"provider": "aws",
			"access_key": "ASIATEMP123",
			"secret_key": "TEMPSECRETXYZ",
			"session_token": "IQoJb3JpZ2luX2Vj...",
			"expires_at": "2025-01-27T10:30:00Z",
			"region": "us-east-1"
            "output": "json"
		}
	]
}
```

## Installation

### Go Install

```
go install github.com/youruser/cloud-swap@latest
```

## Usage

### Add a credential (“Add credit”)

```
cloud-swap add
```

### Add a temporary credential

```
cloud-swap add --temp
```

## List profiles

```
cloud-swap list
```

## Switch profiles

```
cloud-swap use work
eval $(cloud-swap use work)
```

## Optional: Fuzzy selector

```
cloud-swap
```

## STS Assume Role (optional)

```
cloud-swap assume admin-role
```

## Roadmap (TODO)

-   Multi-cloud support
-   Encrypted credential storage
-   Auto-refresh for SSO
-   Full TUI
-   Plugins

## 🤝 Contributing

Pull requests and issues are welcome!

## License

MIT License.
