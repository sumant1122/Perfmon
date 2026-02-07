# Contributing to Perfmon

Thank you for your interest in contributing to Perfmon! We welcome contributions from everyone.

## Getting Started

1.  **Fork the repository** on GitHub.
2.  **Clone your fork** locally.
3.  **Create a branch** for your feature or bug fix.

```bash
git checkout -b feature/my-new-feature
```

## Development

Perfmon is written in Go. You will need Go 1.22 or later installed.

### Build and Run

We use a `Makefile` to simplify common tasks.

*   **Run locally**:
    ```bash
    make run
    ```
*   **Build binary**:
    ```bash
    make build
    ```
*   **Run tests**:
    ```bash
    make test
    ```

### Code Style

Please ensure your code is formatted with `gofmt`. We also recommend running `go vet` before submitting.

## Submitting a Pull Request

1.  Push your branch to your fork.
2.  Open a Pull Request against the `main` branch.
3.  Describe your changes and link to any relevant issues.
4.  Ensure the CI checks pass.

## License

By contributing, you agree that your contributions will be licensed under the project's [LICENSE](./LICENSE).
