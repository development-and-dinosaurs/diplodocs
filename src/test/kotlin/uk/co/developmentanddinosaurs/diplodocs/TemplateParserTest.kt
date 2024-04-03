package uk.co.developmentanddinosaurs.diplodocs

import io.kotest.core.spec.style.StringSpec
import io.kotest.matchers.shouldBe

class TemplateParserTest : StringSpec({

    val parser = TemplateParser()

    "parse should return a single text node with the entire template string" {
        // Arrange
        val template = "Hello, world!"

        // Act
        val templateNodes = parser.parse(template)

        // Assert
        templateNodes.size shouldBe 1
        templateNodes.first() shouldBe TextNode("Hello, world!")
    }
})
