# After Cloning

A few ways to use ghorg after cloning repos. Please add yours below if you have one!

## On Mac

> Dump all test.sh files from ghorg dir into a results file

```
find $HOME/ghorg -name "test.sh" -exec cat {} \; > results
```

> Find any use of gcloud in a file called test.sh

```
find $HOME/ghorg -name "test.sh" | xargs grep -i "gcloud"
```

> Sort cloned repos by size

```
# cd into a clone dir
du -d 1 . | sort -nr | cut -f2- | xargs du -hs
```

## On PC

> Help Wanted

## On Linux

> Help Wanted
