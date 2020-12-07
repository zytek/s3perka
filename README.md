<img src="https://cdn.onlinewebfonts.com/svg/img_535582.png" width="128" height="128">

# s3perka - copy S3 objects between different AWS partitions (China, GovCloud) or accounts

Use this tool if you need to copy S3 objects to/from AWS China or GovCloud, which
are **different AWS partitions** and hence you **cannot** use `aws s3 sync`

You can only use this to copy between different AWS accounts if you can't setup proper
IAM policies on source/destination buckets to use `aws s3 sync/cp`

* supports massive parallelism (100+ simultaneus copies) as this has proven speeding up copying especially to AWS China
* only copies files that do not exist on destination bucket or whose size differs
* does not delete files on destination bucket (think: `aws s3 sync --delete`, this is *unsupported*)


# Example output

```
2020/12/07 15:30:04 [rsg-source->dest-bucket-china] destination: found 32711 objects ( 129 GB )
2020/12/07 15:30:06 [rsg-source->dest-bucket-china] source: found 32712 objects ( 129 GB )
2020/12/07 15:30:06 [rsg-source->dest-bucket-china] starting, must copy 1 new objects ( 0 GB )
2020/12/07 15:30:06 All done
```

# Installation

1. Clone the repo
```
cd ~/git
git clone https://github.com/zytek/s3perka
cd s3perka
```
2. Build, using go 1.13+ for module support

```go build```

This will produce `s3perka` binary.

3. To cross-compile use something like this:
```GOOS=linux GOARCH=amd64 go build```

# Usage

`s3perka` uses *AWS CLI profiles* to get credentials for source and destination AWS accounts. Configure AWS CLI profiles first
and then select proper profile. Empty profile will fallback to using default credentials chain.

Sample `config.toml` file:
```
source.bucket = "BUCKET_SOURCE"
source.region = "eu-west-1"
source.prefix = "somePrefix/"
# Empty profile "" will fallback to default AWS Credentials chain including EC2 Instance Role etc
source.profile = "sourceProf"
destination.bucket = "BUCKET_DEST"
destination.region = "cn-north-1"
destination.prefix = ""
destination.profile = "destProf"
# Default 10, but I recommend using 50-200 especialy when running on EC2
parallel=50
```

# Caveats

* `s3perka` downloads files on disk and then uploads them. Currently there is no "on-the-fly" (in-memory) copy mode
  * make sure you have enough disk space for largest file to fit
  * make sure youd HDD can handle such parallelism level (this can be tweaked on linux system)
