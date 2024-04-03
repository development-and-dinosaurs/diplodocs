package uk.co.developmentanddinosaurs.diplodocs

import io.kotest.core.spec.style.StringSpec
import io.kotest.matchers.shouldBe

class TemplateProcessorTest : StringSpec({
    val processor = TemplateProcessor()

    "processes TextNode correctly" {
        // Arrange
        val textNode = TextNode("Hello, world!")

        // Act
        val result = processor.visit(textNode)

        // Assert
        result shouldBe "Hello, world!"
    }
})
