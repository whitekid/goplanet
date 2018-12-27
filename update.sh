#!/bin/sh
make build
git checkout -f gh-pages
git checkout master -- feeds.ini
bin/goplanet > index.xml
git diff index.xml
git rm -f feeds.ini
git add index.xml feeds.ini
git commit --amend -C HEAD
git push -f
git checkout -f master
