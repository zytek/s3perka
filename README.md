<img src="https://cdn.onlinewebfonts.com/svg/img_535582.png" width="128" height="128">

# s3perka - copy S3 objects between different AWS partitions (China, GovCloud)

Use this tool if you need to copy S3 objects to/from AWS China or GovCloud, which 
are **different AWS partitions** and hence you **cannot** use `aws s3 sync`

* supports massive parallelism (dozens or 100+ simultaneus copies) as this has proven speeding up copying especially to AWS China
* only copies files from source, that do not exist or whose size on destination differs
* does not delete files on destination bucket

# Installation

1. Clone the repo
```
cd ~/git
git clone https://github.com/zytek/s3perka
cd s3perka
```
2.  Build, using go 1.13+ for module support

```go build```

This will produce `s3perka` binary.

# Usage

`s3perka` uses *AWS CLI profiles* to get credentials for source and destination AWS accounts. 

Sample `config.toml` file:
```
source.bucket = "BUCKET_SOURCE"
source.region = "eu-west-1"
source.prefix = "somePrefix/"
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
  * make sure youd HDD can handle such parallelism level 
