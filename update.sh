#!/bin/sh -e
gmake build
git checkout -f gh-pages
git checkout master -- feeds.yaml index.tmpl
bin/goplanet update
git rm -f feeds.yaml index.tmpl
git add index.html golang.xml k8s.xml
git commit --amend -C HEAD
git push -f
git checkout -f master
git gc
