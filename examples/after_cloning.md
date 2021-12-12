# After Cloning

A few ways to use ghorg after cloning repos. Please add yours below if you have one!

> Dump all test.sh files from ghorg dir into a results file

```
find $HOME/Desktop/ghorg -name "test.sh" -exec cat {} \; > results
```

> Find any use of gcloud in a file called test.sh

```
find $HOME/Desktop/ghorg -name "test.sh" | xargs grep -i "gcloud"
```

> Sort cloned repos by size

```
# cd into a clone dir
du -d 1 . | sort -nr | cut -f2- | xargs du -hs
```
