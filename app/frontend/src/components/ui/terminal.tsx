
import * as React from "react";
import { cn } from "@/lib/utils";

interface TerminalProps extends React.HTMLAttributes<HTMLDivElement> {
  children: React.ReactNode;
}

export function Terminal({ children, className, ...props }: TerminalProps) {
  return (
    <div
      className={cn(
        "terminal-output rounded-md",
        className
      )}
      {...props}
    >
      {children}
    </div>
  );
}

interface TerminalLineProps extends React.HTMLAttributes<HTMLDivElement> {
  children: React.ReactNode;
  variant?: "default" | "success" | "error" | "warning";
}

export function TerminalLine({ 
  children, 
  variant = "default", 
  className, 
  ...props 
}: TerminalLineProps) {
  return (
    <div
      className={cn(
        variant === "success" && "text-terminal-success",
        variant === "error" && "text-terminal-error",
        variant === "warning" && "text-terminal-warning",
        className
      )}
      {...props}
    >
      {children}
    </div>
  );
}
