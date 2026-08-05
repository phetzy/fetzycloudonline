# Bootstrap: state bucket

Run this once, before anything else in `terraform/`. It creates the S3
bucket that the main config's backend block uses for remote state.

Because it creates that bucket, this config cannot itself store its state
in it — chicken and egg. It uses local state instead, and the resulting
`terraform.tfstate` (and `.terraform/`) are gitignored. Do not try to move
this config onto a remote backend later; it is meant to stay local and be
run rarely, by hand.

## Usage

```bash
cd terraform/bootstrap
tofu init
tofu apply
```

Note the `state_bucket` output. That bucket name is what the main config
needs in its backend block at `tofu init` time.

## `prevent_destroy`

The bucket has `lifecycle { prevent_destroy = true }`. State is the one
thing here that cannot be rebuilt from the repository, so this config
refuses to let a `tofu destroy` (accidental or otherwise) take it out.
