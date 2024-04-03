package uk.co.developmentanddinosaurs.diplodocs

import io.kotest.core.spec.style.StringSpec
import io.kotest.matchers.shouldBe

class TemplateEngineTest : StringSpec({

    val engine = TemplateEngine()
    val mockVisitor =
        object : TemplateVisitor {
            override fun visit(textNode: TextNode): String {
                return textNode.text
            }
        }

    "processes template nodes correctly" {
        // Arrange
        val textNode1 = TextNode("Hello, ")
        val textNode2 = TextNode("world!")

        // Act
        val result = engine.processTemplate(listOf(textNode1, textNode2), mockVisitor)

        // Assert
        result shouldBe "Hello, world!"
    }
})
