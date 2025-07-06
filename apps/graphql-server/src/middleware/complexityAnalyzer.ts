 export class ComplexityAnalyzer {
  private cache = new Map<string, number>();

  analyze(query: string, variables?: any): number {
    // Create cache key
    const cacheKey = this.createCacheKey(query, variables);
    
    if (this.cache.has(cacheKey)) {
      return this.cache.get(cacheKey)!;
    }

    const complexity = this.calculateComplexity(query);
    this.cache.set(cacheKey, complexity);
    
    return complexity;
  }

  private createCacheKey(query: string, variables?: any): string {
    return `${query}-${JSON.stringify(variables || {})}`;
  }

  private calculateComplexity(query: string): number {
    let complexity = 0;
    
    // Basic complexity scoring (we'll enhance this later)
    // Count various GraphQL operations
    
    // Field selections
    const fieldMatches = query.match(/\w+/g) || [];
    complexity += fieldMatches.length;
    
    // Nested queries increase complexity exponentially
    const braceDepth = this.calculateNestingDepth(query);
    complexity += Math.pow(braceDepth, 2);
    
    // Specific high-complexity operations
    if (query.includes('allUsersWithPostsAndComments')) complexity += 500;
    if (query.includes('userAnalytics')) complexity += 300;
    if (query.includes('systemAnalytics')) complexity += 800;
    if (query.includes('userWithAllData')) complexity += 200;
    if (query.includes('allUsersWithPosts')) complexity += 150;
    
    // Count list fields (arrays increase complexity)
    const listFields = query.match(/\b(posts|comments|users)\b/g) || [];
    complexity += listFields.length * 10;
    
    // Count nested object selections
    const nestedSelections = query.match(/\{\s*\w+\s*\{/g) || [];
    complexity += nestedSelections.length * 20;
    
    return Math.max(complexity, 1); // Minimum complexity of 1
  }

  private calculateNestingDepth(query: string): number {
    let depth = 0;
    let maxDepth = 0;
    
    for (const char of query) {
      if (char === '{') {
        depth++;
        maxDepth = Math.max(maxDepth, depth);
      } else if (char === '}') {
        depth--;
      }
    }
    
    return maxDepth;
  }

  getStats() {
    return {
      cacheSize: this.cache.size,
      cachedQueries: Array.from(this.cache.entries()).map(([query, complexity]) => ({
        query: query.substring(0, 100) + (query.length > 100 ? '...' : ''),
        complexity
      }))
    };
  }

  clearCache() {
    this.cache.clear();
  }
}
