import { Home, Settings, Clock } from "lucide-react";
import { useLocation, Link } from "react-router-dom";
import { useCategories } from "@/hooks/useApi";

import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarTrigger,
} from "@/components/ui/sidebar";

export function AppSidebar() {
  const location = useLocation();
  const { data: categories = [] } = useCategories();
  
  const isActive = (path: string) => {
    return location.pathname === path;
  };

  return (
    <Sidebar>
      <SidebarHeader className="flex items-center px-4 py-2">
        <div className="flex items-center space-x-2">
          <div className="h-8 w-8 rounded-full bg-accent flex items-center justify-center">
            <svg 
              className="h-5 w-5 text-accent-foreground" 
              viewBox="0 0 24 24" 
              fill="none" 
              stroke="currentColor" 
              strokeWidth="2" 
              strokeLinecap="round" 
              strokeLinejoin="round"
            >
              <path d="M16 8a4 4 0 1 1-8 0 4 4 0 0 1 8 0z" />
              <path d="m2 16 2 3h16l2-3" />
              <path d="M9.155 18.39 8 22" />
              <path d="M15 18.5 16 22" />
              <path d="M12 17v5" />
            </svg>
          </div>
          <div className="font-bold text-lg">Developer Loop</div>
        </div>
        <SidebarTrigger className="ml-auto md:hidden" />
      </SidebarHeader>
      
      <SidebarContent>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton asChild isActive={isActive("/")}>
              <Link to="/">
                <Home className="h-5 w-5" />
                <span>Home</span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>

          <SidebarMenuItem>
            <SidebarMenuButton asChild isActive={isActive("/recent")}>
              <Link to="/recent">
                <Clock className="h-5 w-5" />
                <span>Recent</span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
          
          {categories.length > 0 && (
            <SidebarGroup>
              <SidebarGroupLabel>Categories</SidebarGroupLabel>
              <SidebarGroupContent>
                {categories.map(({ category, count }) => (
                  <SidebarMenuItem key={category}>
                    <SidebarMenuButton asChild isActive={isActive(`/category/${category}`)}>
                      <Link to={`/category/${encodeURIComponent(category)}`}>
                        <span>{category || 'Uncategorized'}</span>
                        <span className="text-muted-foreground text-sm">{count}</span>
                      </Link>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarGroupContent>
            </SidebarGroup>
          )}
        </SidebarMenu>
      </SidebarContent>

      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton asChild isActive={isActive("/settings")}>
              <Link to="/settings">
                <Settings className="h-5 w-5" />
                <span>Settings</span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
    </Sidebar>
  );
}
