# Welcome to Diplodocs 🦕

> **Diplodocs** is an opinionated, batteries-included documentation engine.
> Colossal power, zero configuration.

---

## Why Diplodocs?

MkDocs has been the standard for Python documentation for years, but getting a modern, polished site usually requires 10 different plugins and 60+ lines of YAML configuration.

Diplodocs replaces the plugin clutter with a single fast binary:

- 🚀 **Zero-Config Defaults:** Point Diplodocs at your `docs/` folder and start.
- ⚡ **Native Modern Markdown:** GitHub callouts, code tabs, and mermaid diagrams work without third-party plugins.
- 📂 **Filesystem Navigation:** Automatic section hierarchies with natural ordering.
- 🤖 **Agent-Ready (`/llms.txt`):** Generates structured context files for AI coding assistants.
- 🔍 **Instant Offline Search:** Pre-indexed fuzzy search out of the box with `Ctrl+K`.

---

## Interactive Feature Preview

### Callouts / Admonitions

> [!NOTE]
> This note rendered with standard GitHub Markdown syntax. No special plugins required.

> [!TIP]
> You can also use standard MkDocs-style `!!! note` syntax! Diplodocs supports both seamlessly.

> [!WARNING]
> Breaking changes or critical warnings grab attention immediately.

### Code Tabs

=== "Go"

    ```go
    package main

    import "fmt"

    func main() {
        fmt.Println("Hello from Diplodocs! 🦕")
    }
    ```

=== "Python"

    ```python
    def main():
        print("Hello from Diplodocs! 🦕")

    if __name__ == "__main__":
        main()
    ```

=== "JavaScript"

    ```javascript
    console.log("Hello from Diplodocs! 🦕");
    ```
