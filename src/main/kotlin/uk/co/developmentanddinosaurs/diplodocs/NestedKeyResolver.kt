package uk.co.developmentanddinosaurs.diplodocs

object NestedKeyResolver {
    fun resolveNestedKey(key: String, data: Map<String, Any>): String {
        val parts = key.split('.')
        var currentValue: Any? = data
        for (part in parts) {
            if (currentValue !is Map<*, *>) {
                return ""
            }
            currentValue = currentValue[part]
        }
        return currentValue?.toString() ?: ""
    }
}
