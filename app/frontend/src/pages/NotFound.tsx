
import { useLocation } from "react-router-dom";
import { useEffect } from "react";
import { Button } from "@/components/ui/button";
import { AppLayout } from "@/components/AppLayout";
import { Home } from "lucide-react";
import { Link } from "react-router-dom";

const NotFound = () => {
  const location = useLocation();

  useEffect(() => {
    console.error(
      "404 Error: User attempted to access non-existent route:",
      location.pathname
    );
  }, [location.pathname]);

  return (
    <AppLayout>
      <div className="min-h-[calc(100vh-theme(spacing.16))] flex flex-col items-center justify-center">
        <div className="text-center space-y-5">
          <div className="text-7xl font-bold text-accent">404</div>
          <h1 className="text-3xl font-bold">Page not found</h1>
          <p className="text-muted-foreground max-w-md mx-auto">
            Sorry, the page you are looking for doesn't exist or has been moved.
          </p>
          <Button asChild className="mt-4">
            <Link to="/">
              <Home className="h-4 w-4 mr-2" />
              Return home
            </Link>
          </Button>
        </div>
      </div>
    </AppLayout>
  );
};

export default NotFound;
