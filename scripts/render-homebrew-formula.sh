#!/usr/bin/env bash

set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: $0 <tag> <sha256>" >&2
  exit 1
fi

tag="$1"
sha256="$2"
version="${tag#v}"

cat <<EOF
class KoBrowser < Formula
  desc "A fast, token-efficient browser for AI agents"
  homepage "https://github.com/libi/ko-browser"
  url "https://github.com/libi/ko-browser/archive/refs/tags/${tag}.tar.gz"
  sha256 "${sha256}"
  license "MIT"
  version "${version}"
  head "https://github.com/libi/ko-browser.git", branch: "main"

  depends_on "go" => :build
  depends_on "pkgconf" => :build
  depends_on "tesseract"

  def install
    ENV["CGO_ENABLED"] = "1"

    ldflags = %W[
      -s -w
      -X github.com/libi/ko-browser/cmd.version=#{version}
      -X github.com/libi/ko-browser/cmd.commit=homebrew
      -X github.com/libi/ko-browser/cmd.date=homebrew
    ]

    system "go", "build", *std_go_args(output: bin/"kbr", ldflags: ldflags), "-tags=ocr", "./cmd/kbr"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/kbr version")
  end
end
EOF
