import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { AppLayout } from "@/components/AppLayout";
import { ScriptExecutionForm } from "@/components/ScriptExecutionForm";
import { ScriptExecutionHistory } from "@/components/ScriptExecutionHistory";
import { ScriptCodeViewer } from "@/components/ScriptCodeViewer";
import { DeleteScriptModal } from "@/components/DeleteScriptModal";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { ArrowLeft, Star, Code, Terminal as TerminalIcon, PlayCircle, EyeOff, Trash2 } from "lucide-react";
import { formatDistanceToNow } from "date-fns";
import { Terminal } from "@/components/ui/terminal";
import { ScriptExecutionInput, ScriptExecution, Script } from "@/types/script";
import { useScript, useScriptExecutions, useExecuteScript, useDeleteScript } from "@/hooks/useApi";
import { useToast } from "@/hooks/use-toast";

const ScriptDetail = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { toast } = useToast();
  const [activeTab, setActiveTab] = useState("run");
  const [isExecuting, setIsExecuting] = useState(false);
  const [currentExecutionOutput, setCurrentExecutionOutput] = useState<string | null>(null);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);

  const { data: script, isLoading: scriptLoading, error: scriptError } = useScript(id || "");
  const { data: executions = [], isLoading: executionsLoading } = useScriptExecutions(id || "");
  const { mutateAsync: executeScript } = useExecuteScript(id || "");
  const { mutateAsync: deleteScript } = useDeleteScript();

  const handleExecute = async (inputs: ScriptExecutionInput[], isIncognito: boolean) => {
    setIsExecuting(true);
    setCurrentExecutionOutput("Starting script execution...\n");
    
    try {
      const output = await executeScript({ inputs, isIncognito });
      setCurrentExecutionOutput(output);
      setActiveTab("output");
      toast({
        title: 'Script executed successfully',
      });
    } catch (error) {
      console.error("Script execution failed:", error);
      toast({
        title: 'Failed to execute script',
        description: error instanceof Error ? error.message : 'An error occurred',
        variant: 'destructive',
      });
    } finally {
      setIsExecuting(false);
    }
  };

  const handleRerun = async (execution: ScriptExecution) => {
    const filteredInputs = execution.inputs.slice(1);
    handleExecute(filteredInputs, execution.incognito);
  };

  const handleDeleteScript = async (removeFile: boolean) => {
    try {
      await deleteScript({ id: id!, removeFile });
      toast({
        title: 'Script deleted successfully',
      });
      navigate("/");
    } catch (error) {
      toast({
        title: 'Failed to delete script',
        description: error instanceof Error ? error.message : 'An error occurred',
        variant: 'destructive',
      });
    }
  };

  if (scriptLoading) {
    return (
      <AppLayout>
        <div className="container py-8">
          <div className="text-center py-12">
            <h2 className="text-2xl font-bold">Loading Script...</h2>
          </div>
        </div>
      </AppLayout>
    );
  }

  if (!script || scriptError) {
    return (
      <AppLayout>
        <div className="container py-8">
          <div className="text-center py-12">
            <h2 className="text-2xl font-bold">Script Not Found</h2>
            <p className="mt-2 text-muted-foreground">
              The script you're looking for doesn't exist or has been removed.
            </p>
            <Button className="mt-4" onClick={() => navigate("/")}>
              Return to Home
            </Button>
          </div>
        </div>
      </AppLayout>
    );
  }

  return (
    <AppLayout>
      <div className="container py-6 space-y-6">
        <div className="flex items-center gap-2">
          <Button variant="ghost" size="icon" onClick={() => navigate(-1)}>
            <ArrowLeft className="h-5 w-5" />
          </Button>
          <h1 className="text-2xl font-bold tracking-tight">
            {script.name}
          </h1>
          <div className="ml-auto flex gap-2">
            <Button variant="outline" size="sm" onClick={() => setIsDeleteModalOpen(true)}>
              <Trash2 className="h-4 w-4 mr-2" />
              Delete
            </Button>
          </div>
        </div>
        
        <div className="flex flex-col gap-2">
          <p className="text-muted-foreground">
            {script.description}
          </p>
          
          <div className="flex flex-wrap gap-1">
            {script.tags?.map((tag) => (
              <Badge key={tag} variant="secondary">
                {tag}
              </Badge>
            ))}
          </div>
          
          <div className="text-sm text-muted-foreground">
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-x-8 gap-y-1">
              <div>Path: <span className="font-mono">{script.path}</span></div>
              {script.author && <div>Author: {script.author}</div>}
              {script.version && <div>Version: {script.version}</div>}
              {script.lastExecuted && (
                <div>Last executed: {formatDistanceToNow(new Date(script.lastExecuted))} ago</div>
              )}
            </div>
          </div>
        </div>
        
        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <TabsList>
            <TabsTrigger value="run" className="flex items-center gap-1">
              <PlayCircle className="h-4 w-4" />
              <span>Run</span>
            </TabsTrigger>
            <TabsTrigger value="output" className="flex items-center gap-1">
              <TerminalIcon className="h-4 w-4" />
              <span>Output</span>
            </TabsTrigger>
            <TabsTrigger value="history" className="flex items-center gap-1">
              <span>History</span>
              {executionsLoading ? null : <Badge variant="secondary" className="ml-1">{executions.length}</Badge>}
            </TabsTrigger>
            <TabsTrigger value="code" className="flex items-center gap-1">
              <Code className="h-4 w-4" />
              <span>Source</span>
            </TabsTrigger>
          </TabsList>
          
          <TabsContent value="run" className="mt-6">
            <div className="max-w-xl">
              <ScriptExecutionForm 
                script={script}
                onExecute={handleExecute}
                isExecuting={isExecuting}
              />
            </div>
          </TabsContent>
          
          <TabsContent value="output" className="mt-6">
            {currentExecutionOutput ? (
              <div className="space-y-4">
                <div className="flex items-center gap-2">
                  <h3 className="text-lg font-medium">Current Execution Output</h3>
                </div>
                <Terminal>
                  {currentExecutionOutput}
                  {isExecuting && <span className="terminal-cursor"></span>}
                </Terminal>
              </div>
            ) : executions.length > 0 ? (
              <div className="space-y-4">
                <div className="flex items-center gap-2">
                  <h3 className="text-lg font-medium">Latest Execution Output</h3>
                  {executions[0].incognito && (
                    <Badge variant="outline" className="gap-1 text-muted-foreground">
                      <EyeOff className="h-3 w-3 mr-1" />
                      Incognito Mode
                    </Badge>
                  )}
                </div>
                <Terminal>
                  {executions[0].incognito ? 
                    "Output hidden (incognito mode enabled)" : 
                    executions[0].output}
                </Terminal>
                {!executions[0].incognito && 
                  <div className="flex justify-end">
                    <Button 
                      variant="outline" 
                      size="sm" 
                      onClick={() => handleRerun(executions[0])}
                      disabled={isExecuting}
                    >
                      <PlayCircle className="h-4 w-4 mr-1" />
                      Rerun with same parameters
                    </Button>
                  </div>
                }
              </div>
            ) : (
              <div className="text-center py-12 text-muted-foreground">
                <div className="mx-auto w-12 h-12 rounded-full bg-muted flex items-center justify-center mb-3">
                  <TerminalIcon className="h-6 w-6" />
                </div>
                <h3 className="text-lg font-medium">No output yet</h3>
                <p className="mt-1">
                  Run the script to see output here
                </p>
              </div>
            )}
          </TabsContent>
          
          <TabsContent value="history" className="mt-6">
            <ScriptExecutionHistory
              executions={executions}
              isLoading={executionsLoading}
              onRerun={handleRerun}
              scriptId={id}
            />
          </TabsContent>
          
          <TabsContent value="code" className="mt-6">
            <ScriptCodeViewer script={script} />
          </TabsContent>
        </Tabs>

        <DeleteScriptModal
          scriptName={script.name}
          isOpen={isDeleteModalOpen}
          onClose={() => setIsDeleteModalOpen(false)}
          onConfirm={handleDeleteScript}
        />
      </div>
    </AppLayout>
  );
};

export default ScriptDetail;
