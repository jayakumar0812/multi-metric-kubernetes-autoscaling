export class ComplexityAnalyzer {
  private cache = new Map<string, { complexity: number; timestamp: Date; hitCount: number }>();

  analyze(query: string, variables?: any): number {
    // Create cache key
    const cacheKey = this.createCacheKey(query, variables);
    
    console.log(`🔍 Analyzing query (key: ${cacheKey.substring(0, 50)}...)`);
    
    if (this.cache.has(cacheKey)) {
      const cached = this.cache.get(cacheKey)!;
      cached.hitCount++;
      cached.timestamp = new Date();
      
      console.log(`💾 CACHE HIT - Returning cached complexity: ${cached.complexity} (hit count: ${cached.hitCount})`);
      return cached.complexity;
    }

    console.log(`🆕 CACHE MISS - Calculating new complexity...`);
    const complexity = this.calculateComplexity(query);
    
    this.cache.set(cacheKey, {
      complexity,
      timestamp: new Date(),
      hitCount: 1
    });
    
    console.log(`✅ Calculated and cached complexity: ${complexity}`);
    return complexity;
  }

  private createCacheKey(query: string, variables?: any): string {
    // Clean up the query to ensure consistent caching
    const normalizedQuery = query.replace(/\s+/g, ' ').trim();
    return `${normalizedQuery}-${JSON.stringify(variables || {})}`;
  }

  private calculateComplexity(query: string): number {
    let complexity = 0;
    
    console.log(`🧮 Calculating complexity for query: ${query.substring(0, 100)}...`);
    
    // Basic complexity scoring
    const fieldMatches = query.match(/\w+/g) || [];
    complexity += fieldMatches.length;
    console.log(`  📊 Field matches: ${fieldMatches.length}`);
    
    // Nested queries increase complexity exponentially
    const braceDepth = this.calculateNestingDepth(query);
    complexity += Math.pow(braceDepth, 2);
    console.log(`  📊 Nesting depth: ${braceDepth}, depth complexity: ${Math.pow(braceDepth, 2)}`);
    
    // Specific high-complexity operations
    let bonusComplexity = 0;
    if (query.includes('allUsersWithPostsAndComments')) bonusComplexity += 500;
    if (query.includes('userAnalytics')) bonusComplexity += 300;
    if (query.includes('systemAnalytics')) bonusComplexity += 800;
    if (query.includes('userWithAllData')) bonusComplexity += 200;
    if (query.includes('allUsersWithPosts')) bonusComplexity += 150;
    
    complexity += bonusComplexity;
    console.log(`  📊 Bonus complexity: ${bonusComplexity}`);
    
    // Count list fields (arrays increase complexity)
    const listFields = query.match(/\b(posts|comments|users)\b/g) || [];
    complexity += listFields.length * 10;
    console.log(`  📊 List fields: ${listFields.length}`);
    
    // Count nested object selections
    const nestedSelections = query.match(/\{\s*\w+\s*\{/g) || [];
    complexity += nestedSelections.length * 20;
    console.log(`  📊 Nested selections: ${nestedSelections.length}`);
    
    const finalComplexity = Math.max(complexity, 1);
    console.log(`  🎯 Final complexity: ${finalComplexity}`);
    
    return finalComplexity;
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
      cachedQueries: Array.from(this.cache.entries()).map(([query, cacheEntry]) => ({
        query: query.substring(0, 100) + (query.length > 100 ? '...' : ''),
        complexity: cacheEntry.complexity,
        hitCount: cacheEntry.hitCount,
        lastUsed: cacheEntry.timestamp
      })),
      lastUpdated: new Date().toISOString()
    };
  }

  clearCache() {
    console.log(`🗑️ Clearing cache (${this.cache.size} entries)`);
    this.cache.clear();
  }
}