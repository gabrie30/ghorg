# After Cloning

A few ways to use ghorg after cloning repos. Please add yours below if you have one!

> dump all test.sh files from ghorg dir into a results file

```
find $HOME/Desktop/ghorg -name "test.sh" -exec cat {} \; > results
```

> find any use of gcloud in a file called test.sh

```
find $HOME/Desktop/ghorg -name "test.sh" | xargs grep -i "gcloud"
```
