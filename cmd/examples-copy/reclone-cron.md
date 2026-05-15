# Reclone Cron Command

The `reclone-cron` command sets up a simple cron job that triggers the reclone command at specified minute intervals indefinitely.

See the [Reclone Command](https://github.com/gabrie30/ghorg#reclone-command) section of the README for details on configuring `reclone.yaml`.

## Usage

```sh
ghorg reclone-cron [flags]
```

## Flags

- `--minutes`: Specify the interval in minutes at which the reclone command will be triggered. Default is every 60 minutes.

## Example

Set up a cron job to trigger the reclone command every day:

```sh
ghorg reclone-cron --minutes 1440
```

## Environment Variables

- `GHORG_CRON_TIMER_MINUTES`: The interval in minutes for the cron job. This can be set via the `--minutes` flag. Default is 60 minutes.
