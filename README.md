# awsq — Query AWS Resources Instantly

A fast, zero-setup CLI to query AWS resources in human-readable format. Think `kubectl get` but for AWS.

## Why?

Instead of:
```bash
aws ec2 describe-instances | jq -r '.Reservations[].Instances[] | select(.State.Name=="running") | [.InstanceId, .InstanceType, (.Tags[]? | select(.Key=="Name") | .Value)] | @tsv'
```

Just type:
```bash
awsq ec2 --filter state=running
```

## Installation

### Download pre-built binary (recommended)

Download the latest release for your platform from the [Releases page](https://github.com/subashsasi/awsq/releases/latest).

```bash
# Linux (amd64)
curl -Lo awsq https://github.com/subashsasi/awsq/releases/latest/download/awsq-linux-amd64
chmod +x awsq
sudo mv awsq /usr/local/bin/

# Linux (arm64)
curl -Lo awsq https://github.com/subashsasi/awsq/releases/latest/download/awsq-linux-arm64
chmod +x awsq
sudo mv awsq /usr/local/bin/

# macOS (Apple Silicon)
curl -Lo awsq https://github.com/subashsasi/awsq/releases/latest/download/awsq-darwin-arm64
chmod +x awsq
sudo mv awsq /usr/local/bin/

# macOS (Intel)
curl -Lo awsq https://github.com/subashsasi/awsq/releases/latest/download/awsq-darwin-amd64
chmod +x awsq
sudo mv awsq /usr/local/bin/

# Windows — download awsq-windows-amd64.exe from the releases page
# and add it to your PATH
```

### Build from source

Requires [Go 1.22+](https://go.dev/dl/).

```bash
git clone https://github.com/subashsasi/awsq.git
cd awsq
go build -o awsq .

# Move to PATH (Linux/macOS)
sudo mv awsq /usr/local/bin/
```

## Usage

```bash
awsq <resource> [flags]
```

### Global Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--region` | `-r` | auto-detect | AWS region |
| `--profile` | `-p` | default | AWS profile from ~/.aws/credentials |
| `--output` | `-o` | `table` | Output format: table, json, csv |
| `--version` | | | Print version |

### Region Resolution

`awsq` automatically detects your AWS region (no need to pass `-r` every time):

1. `--region` / `-r` flag (highest priority)
2. `AWS_DEFAULT_REGION` environment variable
3. `AWS_REGION` environment variable
4. `~/.aws/config` default region (respects `AWS_PROFILE`)

If no region is found, `awsq` will ask you to specify one.

### Profile Support

Use `--profile` / `-p` to switch between AWS accounts/credentials:

```bash
awsq ec2                       # uses [default] profile
awsq ec2 -p prod              # uses [prod] profile
awsq ec2 --profile staging    # uses [staging] profile
```

Works with named profiles in `~/.aws/credentials`:

```ini
[default]
aws_access_key_id = AKIA...
aws_secret_access_key = ...

[prod]
aws_access_key_id = AKIA...
aws_secret_access_key = ...
```

Also works with SSO profiles in `~/.aws/config`:

```bash
aws sso login --profile dev-sso    # authenticate first
awsq ec2 -p dev-sso               # then query using that profile
```

## Supported Resources

### EC2 Instances

```bash
awsq ec2                                    # All instances
awsq ec2 --filter state=running             # Only running
awsq ec2 --filter state=running,type=t3.micro
awsq ec2 --filter name=web-server           # Filter by Name tag
awsq ec2 -r us-west-2                       # Different region
awsq ec2 -o json                            # JSON output
```

Output:
```
ID                    NAME          TYPE        STATE    PRIVATE_IP     PUBLIC_IP     AZ
--------------------  ----------    ----------  -------  -----------    ----------    -----------
i-0abc123def456789    web-server    t3.micro    running  10.0.1.50      54.1.2.3      us-east-1a
i-0def789abc012345    api-server    t3.large    running  10.0.2.100     -             us-east-1b

(2 instances)
```

### ECS Services

```bash
awsq ecs --cluster prod-cluster
awsq ecs -c default
```

Output:
```
SERVICE             STATUS   DESIRED  RUNNING  PENDING  TASK_DEF
------------------  ------   -------  -------  -------  --------------------------
payment-service     ACTIVE   3        3        0        payment-service:42
auth-service        ACTIVE   2        2        0        auth-service:18

(2 services in cluster 'prod-cluster')
```

### ECS Tasks

```bash
awsq ecs-tasks --cluster prod --service payment-service
awsq ecs-tasks -c prod -s auth-service
```

### Security Groups

```bash
awsq sg
awsq sg --filter vpc=vpc-123abc
```

Output:
```
ID            NAME          VPC           INBOUND_RULES  OUTBOUND_RULES  DESCRIPTION
-----------   ----------    -----------   -------------  --------------  ----------------
sg-0abc123    web-alb-sg    vpc-111aaa    3              1               ALB security group
sg-0def456    app-sg        vpc-111aaa    1              1               App instances

(2 security groups)
```

### ALB (Load Balancers)

```bash
awsq alb
awsq alb -r us-west-2
```

### RDS Databases

```bash
awsq rds
awsq rds -o json
```

Output:
```
ID              ENGINE    VERSION  CLASS          STATUS     MULTI_AZ  STORAGE_GB  ENDPOINT
--------------  --------  -------  -----------    ---------  --------  ----------  ---------------------------
prod-db         postgres  15.3     db.r6g.xlarge  available  Yes       200         prod-db.xxx.us-east-1.rds.amazonaws.com:5432

(1 databases)
```

### Lambda Functions

```bash
awsq lambda
awsq lambda -r us-west-2
```

### VPCs

```bash
awsq vpc
```

Output:
```
VPC_ID        NAME             CIDR           STATE      DEFAULT  SUBNETS
-----------   --------------   -------------- ---------  -------  -------
vpc-111aaa    prod-vpc         10.0.0.0/16    available  No       6
vpc-222bbb    default          172.31.0.0/16  available  Yes      3

(2 VPCs)
```

### Secrets Manager

```bash
awsq secrets
awsq secrets -o json
```

## Output Formats

```bash
awsq ec2 -o table    # Default: human-readable table
awsq ec2 -o json     # JSON (pipe to jq for further processing)
awsq ec2 -o csv      # CSV (import to spreadsheet)
```

## AWS Authentication

awsq uses the standard AWS credential chain:
1. Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
2. Shared credentials file (`~/.aws/credentials`)
3. IAM role (if running on EC2/ECS/Lambda)
4. SSO (`aws sso login` then awsq works)

## Project Structure

```
awsq/
├── main.go                # Entry point
├── go.mod                 # Dependencies
├── cmd/
│   ├── root.go            # Root command, global flags
│   ├── ec2.go             # awsq ec2
│   ├── ecs.go             # awsq ecs
│   ├── ecs_tasks.go       # awsq ecs-tasks
│   ├── sg.go              # awsq sg
│   ├── alb.go             # awsq alb
│   ├── rds.go             # awsq rds
│   ├── lambda.go          # awsq lambda
│   ├── vpc.go             # awsq vpc
│   ├── s3.go              # awsq s3
│   └── secrets.go         # awsq secrets
├── pkg/
│   └── formatter/
│       └── formatter.go   # Table, JSON, CSV output
└── README.md
```

## Roadmap

- [ ] `awsq ecr` — ECR repositories and image scan results
- [ ] `awsq iam-roles` — IAM roles with last used date
- [ ] `awsq route53` — DNS records
- [ ] `awsq ebs` — EBS volumes (attached/unattached)
- [ ] `awsq logs` — CloudWatch Logs query
- [ ] `awsq costs` — Current month cost breakdown
- [ ] Multi-account support (`--profile`)
- [ ] Caching (avoid repeated API calls)
- [ ] Watch mode (`awsq ec2 --watch`)

## License

MIT
