## seldon pipeline infer

run inference on a pipeline

### Synopsis

call a pipeline with a given input and get a prediction

```
seldon pipeline infer <pipelineName> (data) [flags]
```

### Options

```
  -f, --file-path string        inference payload file
      --header stringArray      add header key=value
  -h, --help                    help for infer
      --inference-host string   seldon inference host (default "0.0.0.0:9000")
      --inference-mode string   inference mode rest or grpc (default "rest")
  -i, --iterations int          inference iterations (default 1)
      --show-headers            show headers
  -s, --sticky-session          use sticky session from last infer (only works with inference to experiments)
```

### Options inherited from parent commands

```
  -r, --show-request    show request
  -o, --show-response   show response (default true)
```

### SEE ALSO

* [seldon pipeline](seldon_pipeline.md)	 - manage pipelines

