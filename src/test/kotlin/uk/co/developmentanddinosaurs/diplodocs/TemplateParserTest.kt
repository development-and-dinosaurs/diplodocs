package uk.co.developmentanddinosaurs.diplodocs

import io.kotest.assertions.throwables.shouldThrow
import io.kotest.core.spec.style.StringSpec
import io.kotest.matchers.collections.shouldContainExactly
import io.kotest.matchers.shouldBe
import io.kotest.matchers.throwable.shouldHaveMessage

class TemplateParserTest : StringSpec({

    val parser = TemplateParser()

    "parses template without placeholders" {
        // Arrange
        val template = "Hello, world!"

        // Act
        val nodes = parser.parse(template)

        // Assert
        nodes shouldBe listOf(TextNode("Hello, world!"))
    }

    "parses template with single placeholder" {
        // Arrange
        val template = "Hello, {{ name }}!"

        // Act
        val nodes = parser.parse(template)

        // Assert
        nodes shouldContainExactly
            listOf(
                TextNode("Hello, "),
                PlaceholderNode("name"),
                TextNode("!"),
            )
    }

    "parses template with no spaces around placeholders" {
        // Arrange
        val template = "{{name}}"

        // Act
        val nodes = parser.parse(template)

        // Assert
        nodes shouldBe listOf(PlaceholderNode("name"))
    }

    "parses template with multiple spaces around placeholders" {
        // Arrange
        val template = "{{  name  }}"

        // Act
        val nodes = parser.parse(template)

        // Assert
        nodes shouldBe listOf(PlaceholderNode("name"))
    }

    "parses template with multiple placeholders" {
        // Arrange
        val template = "{{ greeting }}, {{ name }}!"

        // Act
        val nodes = parser.parse(template)

        // Assert
        nodes shouldContainExactly
            listOf(
                PlaceholderNode("greeting"),
                TextNode(", "),
                PlaceholderNode("name"),
                TextNode("!"),
            )
    }

    "parses template with nested placeholder values" {
        // Arrange
        val template = "{{ system.greeting }}, {{ user.name }}!"

        // Act
        val nodes = parser.parse(template)

        // Assert
        nodes shouldContainExactly
            listOf(
                PlaceholderNode("system.greeting"),
                TextNode(", "),
                PlaceholderNode("user.name"),
                TextNode("!"),
            )
    }

    "throws exception for template with invalid placeholder syntax" {
        // Arrange
        val template = "Hello, {{ name !}}!"

        // Act & Assert
        val exception =
            shouldThrow<IllegalArgumentException> {
                parser.parse(template)
            }

        // Assert
        exception shouldHaveMessage "Invalid placeholder syntax: '{{ name ! }}'"
    }
})
