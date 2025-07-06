import { ApolloServer } from '@apollo/server';
import { startStandaloneServer } from '@apollo/server/standalone';
import { typeDefs } from './schema';
import { resolvers } from './resolvers';
import { ComplexityAnalyzer } from './middleware/complexityAnalyzer';

interface Context {
  complexity?: number;
  startTime?: number;
}

async function startServer() {
  // Initialize complexity analyzer
  const complexityAnalyzer = new ComplexityAnalyzer();

  const server = new ApolloServer<Context>({
    typeDefs,
    resolvers,
    plugins: [
      // Complexity analysis plugin
      {
        async requestDidStart() {
          return {
            async didResolveOperation(requestContext) {
              const complexity = complexityAnalyzer.analyze(
                requestContext.request.query || '',
                requestContext.request.variables
              );
              
              // Store complexity in context
              requestContext.contextValue.complexity = complexity;
              
              console.log(`[${new Date().toISOString()}] Query complexity: ${complexity}`);
              
              // Add to response headers
              requestContext.response.http.headers.set(
                'x-query-complexity', 
                complexity.toString()
              );
            },
          };
        },
      },
    ],
  });

  const PORT = Number(process.env.GRAPHQL_PORT) || 4000;

  const { url } = await startStandaloneServer(server, {
    listen: { port: PORT },
    context: async ({ req, res }) => ({
      startTime: Date.now(),
    }),
  });

  console.log(`🚀 GraphQL Server ready at ${url}`);
  console.log(`📊 Visit ${url} to access Apollo Studio`);
  console.log(`🔍 Try these test queries:`);
  console.log(`   Simple:  query { user(id: "1") { name email } }`);
  console.log(`   Complex: query { systemAnalytics { totalUsers topUsers { name } } }`);
  
  // Log complexity analyzer stats periodically
  setInterval(() => {
    const stats = complexityAnalyzer.getStats();
    if (stats.cacheSize > 0) {
      console.log(`📈 Complexity cache has ${stats.cacheSize} queries`);
    }
  }, 30000);
}

// Graceful shutdown
process.on('SIGTERM', () => {
  console.log('SIGTERM received, shutting down gracefully');
  process.exit(0);
});

process.on('SIGINT', () => {
  console.log('SIGINT received, shutting down gracefully');
  process.exit(0);
});

// Start the server
startServer().catch((error) => {
  console.error('Failed to start server:', error);
  process.exit(1);
});