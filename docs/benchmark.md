Donation's Benchmarks
=====================

You can see donations's benchmarks in our [CI server](https://travis-ci.org/topfreegames/donations/) as they get run with every build.

## Creating the performance database

To create a new database for running your benchmarks, just run:

```
$ make drop-perf migrate-perf
```

## Running Benchmarks

If you want to run your own benchmarks, just download the project, and run:

```
$ make run-test-donations run-perf
```

## Generating test data

If you want to run your perf tests against a database with more volume of data, just run this command prior to running the above one:

```
$ make drop-perf migrate-perf db-perf
```

**Warning**: This will take a long time running (around 30m).

## Results

The results should be similar to these:

```
```
