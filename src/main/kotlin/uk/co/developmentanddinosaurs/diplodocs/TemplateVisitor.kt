package uk.co.developmentanddinosaurs.diplodocs

/**
 * Represents a visitor for traversing and processing nodes in a template.
 *
 * This interface defines methods for visiting different types of nodes in a template.
 * Implementations of this interface can define custom behavior for processing each type of node.
 */
interface TemplateVisitor {
    /**
     * Visits a text node in the template.
     *
     * @param node The text node to visit.
     * @return The result of processing the text node.
     */
    fun visit(node: TextNode): String

    /**
     * Visits a placeholder node in the template.
     *
     * @param node The placeholder node to visit.
     * @return The result of processing the placeholder node.
     */
    fun visit(node: PlaceholderNode): String
}
