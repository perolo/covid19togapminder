#!/bin/bash
# Batchfile to show how the exmaples files are generated

go build covid19togapminder.go

mv covid19togapminder build/linux/.

build/linux/covid19togapminder -h

echo ""
echo ""

build/linux/covid19togapminder -dir=/home/perolo/src/COVID-19/csse_covid_19_data/csse_covid_19_time_series -out=example/gapminder.csv -US=false -subset=Country -pop=true

build/linux/covid19togapminder -dir=/home/perolo/src/COVID-19/csse_covid_19_data/csse_covid_19_time_series -out=example/gapminderglob.csv -US=false

build/linux/covid19togapminder -dir=/home/perolo/src/COVID-19/csse_covid_19_data/csse_covid_19_time_series -out=example/gapminderUS.csv -US=true

build/linux/covid19togapminder -dir=/home/perolo/src/COVID-19/csse_covid_19_data/csse_covid_19_time_series -out=example/gapminderUSYork.csv -US=true -subset="New York"

