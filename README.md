# garbledreplacer

![GitHub](https://img.shields.io/github/license/tomtwinkle/garbledreplacer)
[![Go Report Card](https://goreportcard.com/badge/github.com/olvrng/ujson?style=flat-square)](https://goreportcard.com/report/github.com/tomtwinkle/garbledreplacer)
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/tomtwinkle/garbledreplacer/build.yaml)
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/tomtwinkle/garbledreplacer/test.yaml?label=test)

## Overview
`transform.Transformer` to replace characters that cannot be converted from `UTF-8` to another character encoding.

`UTF-8` から別の文字コードに変換する際、変換出来ない文字を別の文字に置き換えるための `transform.Transformer`。

## Motivation

Go標準の `japanese.ShiftJIS` 等のEncoderでは変換出来ない文字が合った場合
`rune not supported by encoding` errorが出てしまい変換ができない。

そのため、Encoderを通す前に変換できない文字を事前に別の文字に置き換える為のTransformerを作成した。

`japanese.ShiftJIS` `japanese.EUCJP` `traditionalchinese.Big5` などのEncoderの前に`transform.Chain`で動作する薄いwrapperとなっている。

## Usage

```golang
const msg = "一二三四🍣五六七八九🍺十拾壱"

var buf bytes.Buffer
w := transform.NewWriter(&buf, garbledreplacer.NewTransformer(japanese.ShiftJIS, '?'))
if _, err := w.Write([]byte([]byte(msg))); err != nil {
    panic(err)
}
if err := w.Close(); err != nil {
    panic(err)
}

fmt.Println(buf.String())
// 一二三四?五六七八九?十拾壱
```
