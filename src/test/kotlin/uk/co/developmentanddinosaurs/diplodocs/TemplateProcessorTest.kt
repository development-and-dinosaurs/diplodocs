package uk.co.developmentanddinosaurs.diplodocs

import io.kotest.core.spec.style.StringSpec
import io.kotest.matchers.shouldBe

class TemplateProcessorTest : StringSpec({
    val processor = TemplateProcessor(mapOf("key" to "value", "nested" to mapOf("key" to "value")))

    "processes TextNode" {
        // Arrange
        val textNode = TextNode("Hello, world!")

        // Act
        val result = processor.visit(textNode)

        // Assert
        result shouldBe "Hello, world!"
    }

    "processes simple placeholder node" {
        // Arrange
        val placeholderNode = PlaceholderNode("key")

        // Act
        val result = processor.visit(placeholderNode)

        // Assert
        result shouldBe "value"
    }

    "processes nested placeholder node" {
        // Arrange
        val placeholderNode = PlaceholderNode("nested.key")

        // Act
        val result = processor.visit(placeholderNode)

        // Assert
        result shouldBe "value"
    }

    "returns empty string when simple placeholder is missing from data" {
        // Arrange
        val placeholderNode = PlaceholderNode("missing")

        // Act
        val result = processor.visit(placeholderNode)

        // Assert
        result shouldBe ""
    }

    "returns empty string when nested placeholder is missing from data" {
        // Arrange
        val placeholderNode = PlaceholderNode("missing.key")

        // Act
        val result = processor.visit(placeholderNode)

        // Assert
        result shouldBe ""
    }
})
