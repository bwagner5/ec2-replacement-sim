# ec2-replacement-sim


## Usage:

```
Usage:
  ec2-replacement-sim [flags]

Flags:
      --capacity-type string       Capacity Type (spot or on-demand) (default "spot")
  -f, --file string                YAML Config File
      --flexibility string         Flexibility Set (regex) (default "^(c|m|r).[a-z0-9]+$")
  -h, --help                       help for ec2-replacement-sim
  -o, --output string              Output mode: [short wide yaml] (default "short")
      --pricing-multiplier float   Pricing Multipler to determine replacement threshold (default 0.5)
      --region string              AWS Region
      --replacement string         Replacement Instance Type
      --verbose                    Verbose output
      --version                    version
```

## Installation:

```
brew install bwagner5/wagner/ec2-replacement-sim
```

Packages, binaries, and archives are published for all major platforms (Mac amd64/arm64 & Linux amd64/arm64):

Debian / Ubuntu:

```
[[ `uname -m` == "aarch64" ]] && ARCH="arm64" || ARCH="amd64"
OS=`uname | tr '[:upper:]' '[:lower:]'`
wget https://github.com/bwagner5/ec2-replacement-sim/releases/download/v0.0.1/ec2-replacement-sim_0.0.1_${OS}_${ARCH}.deb
dpkg --install ec2-replacement-sim_0.0.2_linux_amd64.deb
ec2-replacement-sim --help
```

RedHat:

```
[[ `uname -m` == "aarch64" ]] && ARCH="arm64" || ARCH="amd64"
OS=`uname | tr '[:upper:]' '[:lower:]'`
rpm -i https://github.com/bwagner5/ec2-replacement-sim/releases/download/v0.0.1/ec2-replacement-sim_0.0.1_${OS}_${ARCH}.rpm
```

Download Binary Directly:

```
[[ `uname -m` == "aarch64" ]] && ARCH="arm64" || ARCH="amd64"
OS=`uname | tr '[:upper:]' '[:lower:]'`
wget -qO- https://github.com/bwagner5/ec2-replacement-sim/releases/download/v0.0.1/ec2-replacement-sim_0.0.1_${OS}_${ARCH}.tar.gz | tar xvz
chmod +x ec2-replacement-sim
```

## Examples: 

```
> ec2-replacement-sim --replacement m5.xlarge --flexibility='^m5.[a-z0-9]+$' --region us-west-2 --verbose
Waiting for pricing data to be pulled
2023/06/19 12:52:59 m5.large ($0.058) was not below the pricing threshold of $0.043:
2023/06/19 12:52:59 m5.xlarge ($0.086) was not below the pricing threshold of $0.043:
2023/06/19 12:52:59 m5.2xlarge ($0.186) was not below the pricing threshold of $0.043:
2023/06/19 12:52:59 m5.4xlarge ($0.450) was not below the pricing threshold of $0.043:
2023/06/19 12:52:59 m5.8xlarge ($0.682) was not below the pricing threshold of $0.043:
2023/06/19 12:52:59 m5.12xlarge ($1.110) was not below the pricing threshold of $0.043:
2023/06/19 12:52:59 m5.16xlarge ($2.051) was not below the pricing threshold of $0.043:
2023/06/19 12:52:59 m5.24xlarge ($2.269) was not below the pricing threshold of $0.043:
2023/06/19 12:52:59 m5.metal ($2.451) was not below the pricing threshold of $0.043:
Replacement Instance Type: m5.xlarge
      Instance Type Price: $0.086
          Threshold Price: $0.043
          Flexibility Set: 9
   Replacement Candidates: 0
```

```
> ec2-replacement-sim --replacement m5.4xlarge --flexibility='^m5.[a-z0-9]+$' --region us-west-2 --verbose
Waiting for pricing data to be pulled
2023/06/19 12:53:42 m5.4xlarge ($0.450) was not below the pricing threshold of $0.225:
2023/06/19 12:53:42 m5.8xlarge ($0.682) was not below the pricing threshold of $0.225:
2023/06/19 12:53:42 m5.12xlarge ($1.110) was not below the pricing threshold of $0.225:
2023/06/19 12:53:42 m5.16xlarge ($2.051) was not below the pricing threshold of $0.225:
2023/06/19 12:53:42 m5.24xlarge ($2.269) was not below the pricing threshold of $0.225:
2023/06/19 12:53:42 m5.metal ($2.451) was not below the pricing threshold of $0.225:
Replacement Instance Type: m5.4xlarge
      Instance Type Price: $0.450
          Threshold Price: $0.225
          Flexibility Set: 9
   Replacement Candidates: 3
  - m5.large
  - m5.xlarge
  - m5.2xlarge
```

```
> ec2-replacement-sim --replacement m5.8xlarge --flexibility='^m5.[a-z0-9]+$' --region us-west-2 --verbose
Waiting for pricing data to be pulled
2023/06/19 12:54:14 m5.4xlarge ($0.450) was not below the pricing threshold of $0.341:
2023/06/19 12:54:14 m5.8xlarge ($0.682) was not below the pricing threshold of $0.341:
2023/06/19 12:54:14 m5.12xlarge ($1.110) was not below the pricing threshold of $0.341:
2023/06/19 12:54:14 m5.16xlarge ($2.051) was not below the pricing threshold of $0.341:
2023/06/19 12:54:14 m5.24xlarge ($2.269) was not below the pricing threshold of $0.341:
2023/06/19 12:54:14 m5.metal ($2.451) was not below the pricing threshold of $0.341:
Replacement Instance Type: m5.8xlarge
      Instance Type Price: $0.682
          Threshold Price: $0.341
          Flexibility Set: 9
   Replacement Candidates: 3
  - m5.large
  - m5.xlarge
  - m5.2xlarge
```
