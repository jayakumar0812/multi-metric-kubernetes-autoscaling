 // Mock data for testing
const users = [
  { id: '1', name: 'John Doe', email: 'john@example.com', createdAt: new Date().toISOString() },
  { id: '2', name: 'Jane Smith', email: 'jane@example.com', createdAt: new Date().toISOString() },
  { id: '3', name: 'Bob Johnson', email: 'bob@example.com', createdAt: new Date().toISOString() },
  { id: '4', name: 'Alice Brown', email: 'alice@example.com', createdAt: new Date().toISOString() },
  { id: '5', name: 'Charlie Wilson', email: 'charlie@example.com', createdAt: new Date().toISOString() },
];

const posts = [
  { id: '1', title: 'First Post', content: 'Content for first post', authorId: '1', tags: ['tech'], createdAt: new Date().toISOString() },
  { id: '2', title: 'Second Post', content: 'Content for second post', authorId: '2', tags: ['programming'], createdAt: new Date().toISOString() },
  { id: '3', title: 'Third Post', content: 'Content for third post', authorId: '1', tags: ['tutorial'], createdAt: new Date().toISOString() },
  { id: '4', title: 'GraphQL Best Practices', content: 'How to use GraphQL effectively', authorId: '3', tags: ['graphql', 'best-practices'], createdAt: new Date().toISOString() },
  { id: '5', title: 'Auto-scaling in Kubernetes', content: 'Understanding Kubernetes auto-scaling', authorId: '2', tags: ['kubernetes', 'scaling'], createdAt: new Date().toISOString() },
];

const comments = [
  { id: '1', content: 'Great post!', authorId: '2', postId: '1', createdAt: new Date().toISOString() },
  { id: '2', content: 'Thanks for sharing', authorId: '3', postId: '1', createdAt: new Date().toISOString() },
  { id: '3', content: 'Very helpful tutorial', authorId: '4', postId: '3', createdAt: new Date().toISOString() },
  { id: '4', content: 'Looking forward to more content', authorId: '5', postId: '2', createdAt: new Date().toISOString() },
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
      if (!user) return null;
      
      // Simulate expensive operation
      const delay = Math.random() * 50; // Random delay to simulate complexity
      return new Promise(resolve => setTimeout(() => resolve(user), delay));
    },
    
    allUsersWithPosts: () => {
      // High complexity - returns all users with their posts
      return users;
    },
    
    // Very high complexity queries
    allUsersWithPostsAndComments: () => {
      // Extremely high complexity - full data dump
      return users;
    },
    
    userAnalytics: (_: any, { userId }: { userId: string }) => {
      const user = users.find(u => u.id === userId);
      if (!user) throw new Error('User not found');
      
      const userPosts = posts.filter(p => p.authorId === userId);
      const userComments = comments.filter(c => c.authorId === userId);
      
      // Simulate complex analytics calculations
      const avgPostLength = userPosts.length > 0 
        ? userPosts.reduce((acc, p) => acc + p.content.length, 0) / userPosts.length 
        : 0;
      
      return {
        user,
        totalPosts: userPosts.length,
        totalComments: userComments.length,
        averagePostLength: avgPostLength,
        mostPopularPost: userPosts[0] || null,
        engagementScore: Math.random() * 100
      };
    },
    
    systemAnalytics: () => {
      // Maximum complexity query
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
      // This creates N+1 queries - increases complexity
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
