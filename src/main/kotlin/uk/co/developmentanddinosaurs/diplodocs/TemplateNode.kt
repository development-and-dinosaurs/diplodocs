package uk.co.developmentanddinosaurs.diplodocs

/**
 * Represents a node in a template.
 *
 * Encapsulates the different types of template nodes.
 */
sealed class TemplateNode {
    /**
     * Accepts the specified visitor.
     * @param visitor The visitor to accept.
     */
    abstract fun accept(visitor: TemplateVisitor): String
}

/**
 * Represents a text node in a template.
 *
 * Text nodes contain raw text content to be included in the template output.
 *
 * @property text The text content of the node.
 */
data class TextNode(val text: String) : TemplateNode() {
    /**
     * Accepts the specified visitor and calls visit with itself.
     * @param visitor The visitor to accept.
     */
    override fun accept(visitor: TemplateVisitor): String {
        return visitor.visit(this)
    }
}
