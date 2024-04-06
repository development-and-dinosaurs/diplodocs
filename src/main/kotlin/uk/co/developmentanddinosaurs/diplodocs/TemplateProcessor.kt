package uk.co.developmentanddinosaurs.diplodocs

/**
 * Implements the Visitor interface to process template nodes.
 * This class defines the behavior for visiting each type of template node.
 */
class TemplateProcessor(private val data: Map<String, Any>) : TemplateVisitor {
    /**
     * Processes a text node.
     *
     * @param textNode The text node to process.
     * @return The processed text content.
     */
    override fun visit(textNode: TextNode): String {
        return textNode.text
    }

    /**
     * Processes a placeholder node.
     *
     * @param placeholderNode The placeholder node to process.
     * @return The processed placeholder content.
     */
    override fun visit(placeholderNode: PlaceholderNode): String {
        return NestedKeyResolver.resolveNestedKey(placeholderNode.placeholder, data)
    }
}
