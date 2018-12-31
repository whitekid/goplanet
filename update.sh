#!/bin/sh
make build
git checkout -f gh-pages
git checkout master -- feeds.ini
bin/goplanet
git rm -f feeds.ini
git add golang.xml
git commit --amend -C HEAD
git push -f
git checkout -f master
