# Contributing to Perfmon

Thank you for your interest in contributing to Perfmon! We welcome contributions from everyone. This document provides guidelines for contributing to the project.

## Table of Contents

- [How Can I Contribute?](#how-can-i-contribute)
  - [Reporting Bugs](#reporting-bugs)
  - [Suggesting Features](#suggesting-features)
  - [Pull Requests](#pull-requests)
- [Development Setup](#development-setup)
- [Style Guide](#style-guide)
- [Commit Messages](#commit-messages)

## How Can I Contribute?

### Reporting Bugs

Bugs are tracked as GitHub issues. When creating a bug report, please include as much detail as possible:
- A clear, descriptive title.
- The version of Perfmon you are using.
- Your operating system and terminal emulator.
- Steps to reproduce the bug.
- Actual vs. expected behavior.

### Suggesting Features

If you have an idea for a feature, please search existing issues to see if it's already been suggested. If not, open a new issue and describe:
- Why the feature would be useful.
- How it should work.
- Any relevant screenshots or mockups.

### Pull Requests

1.  **Fork** the repository.
2.  **Clone** your fork locally.
3.  Create a new **branch** for your changes.
    ```bash
    git checkout -b feature/your-feature-name
    ```
4.  Make your changes and write **tests** if applicable.
5.  Ensure all tests pass and the code is formatted correctly.
6.  **Push** your branch to your fork.
7.  Open a **Pull Request** against the `main` branch.

## Development Setup

### Prerequisites

- [Go](https://go.dev/dl/) 1.22 or later.
- `make` (optional, for using the Makefile).

### Build and Run

- **Run locally**:
  ```bash
  make run
  # or
  go run .
  ```
- **Build binary**:
  ```bash
  make build
  # or
  go build -o perfmon
  ```
- **Run tests**:
  ```bash
  make test
  # or
  go test ./...
  ```

## Style Guide

- Follow standard Go conventions and idioms.
- Format your code with `gofmt`.
- Use descriptive names for variables and functions.
- Keep functions small and focused on a single task.

## Commit Messages

- Use the imperative mood ("Add feature" not "Added feature").
- Keep the first line short (under 50 characters).
- Provide more detail in the body if necessary, separated by a blank line.

---

By contributing, you agree that your contributions will be licensed under the project's [MIT License](./LICENSE).
