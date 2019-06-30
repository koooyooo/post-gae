$ gcloud app deploy
$ gcloud app logs tail -s default
$ gcloud app browse

$ x: goapp serve
$ dev_appserver.py .



#### goapp install
```
$ gcloud components update

# cloud SDK のパスを確認
$ which gcloud

# PATHを追加し実行権限を追加
$ export PATH=$PATH:${google-cloud-sdk}/platform/google_appengine
$ chmod +x ${google-cloud-sdk}/platform/google_appengine/goapp
```

which gcloud を元に以下のパスを通す
```
export PATH=$PATH:${google-cloud-sdk}/platform/google_appengine
```

#### for using go mod init
```
export GO111MODULE=on
```