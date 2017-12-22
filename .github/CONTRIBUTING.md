# How to contribute

This document contains information about the contribution guidelines. Please refer to the following sections to learn more
about:

  * [Reporting Issues](#reporting-issues)
  * [Pull Request Submission](#pull-request-submission)
  * [Contributing Code](#contributing-code)
    * [Committing Code](#committing-code)
    * [Coding Standards](#coding-standards)
  * [Documentation](#documentation)
  
## Reporting Issues

Issues reported are more than welcome as they help us understand what's broken, what features are missing or even ways to
enhance the product. In order to keep the process of reporting an issue organised we encourage contributors to follow the steps below:

1. Double check the [list of issues](https://github.com/dikhan/terraform-provider-api/issues) already opened and
if the issue is not reported feel free to create a new one.
2. Providing enough context and details about the issue is critical for people to understand the problem. Thus, an
[issue template](ISSUE_TEMPLATE.md) is provided to help filling out the blanks. 

## Pull Request Submission

Pull requests are appreciated, please follow the steps below when creating a new pull request:

1. Read [Creating a Pull Request](https://help.github.com/articles/creating-a-pull-request/) to learn more about how to
submit a Pull Request in Github.
2. A [Pull request template](PULL_REQUEST_TEMPLATE.md) is provided to help populating the PR. Please provide a clear 
description about the change requested.

## Contributing Code 

The best way to contribute to this repository is by creating a [fork](https://help.github.com/articles/fork-a-repo/)
and follow the below: 

- Create a feature branch prefixed with the type of change (bugfix/feature):
    - Bug fix: `git checkout -b **bugfix**/my-bug-fix`
    - Feature requests: `git checkout -b **feature**/my-new-feature`
- Commit your changes: `git commit -am 'Add some feature'`
- Push to the branch: `git push origin <branch-name>`
- Submit a pull request following the [Pull request guidelines](#pull-request-submissio)

## Licensing

Code on the Terraform Provider API GitHub repository is licensed under the terms of the Apache 2.0 license: https://www.apache.org/licenses/LICENSE-2.0 and [https://github.com/dikhan/terraform-provider-api/blob/master/LICENSE]. This license ensures a balance between openness and allowing you to use the code with minimal requirements.

## Licensing Code Contributions

All code that you write yourself and contribute to the Terraform Provider API GitHub repository must be licensed under the Apache 2.0 license. If you wrote the code as part of work for someone else (like a company), you must ensure that you have the proper rights and permissions to contribute the code under the terms of the Apache 2.0 license.

If you want to contribute any code that you did not write yourself (like pre-existing open source code), either alone or in combination with code that you did write, that code must be available under the Apache 2.0, BSD, or MIT license.

If you want to contribute code to the Terraform Provider API GitHub repository that is under any different license terms than specified above, please contact [dkhanram@cisco.com] to request a review.

### Committing Code

Commits are very important to understand the history of the repo changes and also if done well greatly help the reviewing 
process. We are committed to keep the repository code organised and therefore expect contributors to follow some
best practises as enlisted below:

- Commit related messages: A commit should contain related changes.
- Commit often: Small commits make it easier for other developers to understand the changes and also make it easier to
roll back the changes if needed.
- Write good commit messages: What is good? Imagine you did not know anything about the change and had to learn more about 
it by reading the commit message - then write the commit message.

### Coding Standards

- Follow the go coding standards as outlined in [Effective go](https://golang.org/doc/effective_go.html)

## Documentation

This repository aims to translate the latest Swagger OpenApi spec with the corresponding configuration in latest
versions of Terraform. These two products are constantly evolving and adding support for new features and therefore
keeping the docs updated is paramount. Any contributor should keep in mind this and update the [How to](../docs/how_to.md) 
accordingly based on new features added or updated.