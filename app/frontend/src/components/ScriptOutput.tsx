import { Terminal } from "@/components/ui/terminal";
import { Loader2 } from "lucide-react";

interface ScriptOutputProps {
  output: string;
  status: 'idle' | 'running' | 'completed' | 'error';
  isExecuting: boolean;
}

export function ScriptOutput({ output, status, isExecuting }: ScriptOutputProps) {
  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        {isExecuting && (
          <Loader2 className="h-4 w-4 animate-spin" />
        )}
        <span className="text-sm font-medium">
          {status === 'running' && 'Running script...'}
          {status === 'completed' && 'Execution completed'}
          {status === 'error' && 'Execution failed'}
          {status === 'idle' && 'No output yet'}
        </span>
      </div>
      
      <Terminal className="h-[500px]">
        <pre className="whitespace-pre-wrap font-mono text-sm">
          {output || 'No output available'}
        </pre>
      </Terminal>
    </div>
  );
} 