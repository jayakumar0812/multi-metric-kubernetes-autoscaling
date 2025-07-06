 // src/schema/index.ts
export const typeDefs = `#graphql
  type User {
    id: ID!
    name: String!
    email: String!
    posts: [Post!]!
    createdAt: String!
  }

  type Post {
    id: ID!
    title: String!
    content: String!
    author: User!
    comments: [Comment!]!
    tags: [String!]!
    createdAt: String!
  }

  type Comment {
    id: ID!
    content: String!
    author: User!
    post: Post!
    createdAt: String!
  }

  type UserAnalytics {
    user: User!
    totalPosts: Int!
    totalComments: Int!
    averagePostLength: Float!
    mostPopularPost: Post
    engagementScore: Float!
  }

  type SystemAnalytics {
    totalUsers: Int!
    totalPosts: Int!
    totalComments: Int!
    averageComplexity: Float!
    topUsers: [User!]!
    recentPosts: [Post!]!
  }

  type Query {
    # Simple queries (complexity: 1-50)
    user(id: ID!): User
    post(id: ID!): Post
    
    # Medium complexity queries (complexity: 50-200)
    users(limit: Int = 10, offset: Int = 0): [User!]!
    posts(limit: Int = 10, offset: Int = 0): [Post!]!
    
    # High complexity queries (complexity: 200-500)
    userWithAllData(id: ID!): User
    allUsersWithPosts: [User!]!
    
    # Very high complexity queries (complexity: 500+)
    allUsersWithPostsAndComments: [User!]!
    userAnalytics(userId: ID!): UserAnalytics!
    systemAnalytics: SystemAnalytics!
  }
`;

// src/resolvers/index.ts
// Mock data for testing
const users = [
  { id: '1', name: 'John Doe', email: 'john@example.com', createdAt: new Date().toISOString() },
  { id: '2', name: 'Jane Smith', email: 'jane@example.com', createdAt: new Date().toISOString() },
  { id: '3', name: 'Bob Johnson', email: 'bob@example.com', createdAt: new Date().toISOString() },
];

const posts = [
  { id: '1', title: 'First Post', content: 'Content for first post', authorId: '1', tags: ['tech'], createdAt: new Date().toISOString() },
  { id: '2', title: 'Second Post', content: 'Content for second post', authorId: '2', tags: ['programming'], createdAt: new Date().toISOString() },
  { id: '3', title: 'Third Post', content: 'Content for third post', authorId: '1', tags: ['tutorial'], createdAt: new Date().toISOString() },
];

const comments = [
  { id: '1', content: 'Great post!', authorId: '2', postId: '1', createdAt: new Date().toISOString() },
  { id: '2', content: 'Thanks for sharing', authorId: '3', postId: '1', createdAt: new Date().toISOString() },
];

export const resolvers = {
  Query: {
    // Simple queries (low complexity)
    user: (_: any, { id }: { id: string }) => {
      return users.find(user => user.id === id);
    },
    
    post: (_: any, { id }: { id: string }) => {
      return posts.find(post => post.id === id);
    },
    
    // Medium complexity queries
    users: (_: any, { limit, offset }: { limit: number; offset: number }) => {
      return users.slice(offset, offset + limit);
    },
    
    posts: (_: any, { limit, offset }: { limit: number; offset: number }) => {
      return posts.slice(offset, offset + limit);
    },
    
    // High complexity queries
    userWithAllData: (_: any, { id }: { id: string }) => {
      const user = users.find(u => u.id === id);
      return user;
    },
    
    allUsersWithPosts: () => {
      return users;
    },
    
    // Very high complexity queries
    allUsersWithPostsAndComments: () => {
      return users;
    },
    
    userAnalytics: (_: any, { userId }: { userId: string }) => {
      const user = users.find(u => u.id === userId);
      if (!user) throw new Error('User not found');
      
      const userPosts = posts.filter(p => p.authorId === userId);
      const userComments = comments.filter(c => c.authorId === userId);
      
      return {
        user,
        totalPosts: userPosts.length,
        totalComments: userComments.length,
        averagePostLength: userPosts.length > 0 
          ? userPosts.reduce((acc, p) => acc + p.content.length, 0) / userPosts.length 
          : 0,
        mostPopularPost: userPosts[0] || null,
        engagementScore: Math.random() * 100
      };
    },
    
    systemAnalytics: () => {
      return {
        totalUsers: users.length,
        totalPosts: posts.length,
        totalComments: comments.length,
        averageComplexity: 150, // Mock value
        topUsers: users.slice(0, 3),
        recentPosts: posts.slice(0, 5),
      };
    },
  },
  
  User: {
    posts: (parent: any) => {
      return posts.filter(post => post.authorId === parent.id);
    },
  },
  
  Post: {
    author: (parent: any) => {
      return users.find(user => user.id === parent.authorId);
    },
    comments: (parent: any) => {
      return comments.filter(comment => comment.postId === parent.id);
    },
  },
  
  Comment: {
    author: (parent: any) => {
      return users.find(user => user.id === parent.authorId);
    },
    post: (parent: any) => {
      return posts.find(post => post.id === parent.postId);
    },
  },
};

// src/middleware/complexityAnalyzer.ts
export class ComplexityAnalyzer {
  private cache = new Map<string, number>();

  analyze(query: string, variables?: any): number {
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
    
    // Basic complexity scoring
    const fieldMatches = query.match(/\w+/g) || [];
    complexity += fieldMatches.length;
    
    // Nested queries increase complexity
    const braceDepth = this.calculateNestingDepth(query);
    complexity += Math.pow(braceDepth, 2);
    
    // Specific high-complexity operations
    if (query.includes('allUsersWithPostsAndComments')) complexity += 500;
    if (query.includes('userAnalytics')) complexity += 300;
    if (query.includes('systemAnalytics')) complexity += 800;
    if (query.includes('userWithAllData')) complexity += 200;
    
    return Math.max(complexity, 1);
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
        query: query.substring(0, 100),
        complexity
      }))
    };
  }
}
