# Miss-matching Cucumber-Groovy Scenario Steps

[![Build Status](https://semaphoreci.com/api/v1/milanaleksic/MissmatchingCucumberGroovyScenarioSteps/branches/master/badge.svg)](https://semaphoreci.com/milanaleksic/MissmatchingCucumberGroovyScenarioSteps)

This utility finds which steps in Cucumber Groovy `.feature` files refer to non-existing step definitions from `.groovy` files.

I wrote it in Golang since:
 1. Cucumber is silently failing and skipping these steps, introducing confusion
 2. running `dry-run` is far too slow in big projects (with >1000 definitions), this app takes ~2 seconds (without any optimizations applied at this time).

## Building, tagging and artifact deployment

This is `#golang` project. I used Go 1.5.

`go get github.com/milanaleksic/MissmatchingCucumberGroovyScenarioSteps` should be enough to get the code and build.

To build project you can execute (this will get from internet all 3rd party utilites needed for deployment: upx, go-upx, github-release):

    make prepare

You can start building project using `make`, even `deploy` to Github (if you have privileges to do that of course).
