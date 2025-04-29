import { useState } from "react";
import { AppLayout } from "@/components/AppLayout";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Separator } from "@/components/ui/separator";
import { Badge } from "@/components/ui/badge";
import { Plus, Save, Trash2, FolderOpen, RefreshCw } from "lucide-react";
import { AppConfig, ExecutionConfig } from "@/types/script";
import { useToast } from "@/hooks/use-toast";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Form, FormField, FormItem, FormLabel, FormControl, FormDescription } from "@/components/ui/form";
import { useForm } from "react-hook-form";
import { useAppConfig, useUpdateAppConfig, useReloadScripts } from "@/hooks/useApi";

const Settings = () => {
  const [newFolder, setNewFolder] = useState("");
  const [newExtension, setNewExtension] = useState("");
  const [newCommand, setNewCommand] = useState("");
  const { toast } = useToast();
  const { data: config, isLoading } = useAppConfig();
  const { mutateAsync: updateConfig } = useUpdateAppConfig();
  const { mutateAsync: reloadScripts } = useReloadScripts();

  const handleSave = async () => {
    if (!config) return;
    
    try {
      await updateConfig(config);
      toast({
        title: "Success",
        description: "Settings saved successfully",
      });
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to save settings",
        variant: "destructive",
      });
    }
  };

  const handleAddFolder = () => {
    if (!newFolder || !config) return;
    
    if (config.scriptFolders.includes(newFolder)) {
      toast({
        title: "Error",
        description: "This folder is already in the list",
        variant: "destructive",
      });
      return;
    }
    
    updateConfig({
      ...config,
      scriptFolders: [...config.scriptFolders, newFolder]
    });
    
    setNewFolder("");
  };

  const handleRemoveFolder = (folder: string) => {
    if (!config) return;
    
    updateConfig({
      ...config,
      scriptFolders: config.scriptFolders.filter(f => f !== folder)
    });
  };

  const handleAddExtension = () => {
    if (!newExtension || !newCommand || !config) return;
    
    if (config.extensionCommands[newExtension]) {
      toast({
        title: "Error",
        description: "This extension is already configured",
        variant: "destructive",
      });
      return;
    }
    
    updateConfig({
      ...config,
      extensionCommands: {
        ...config.extensionCommands,
        [newExtension]: newCommand
      }
    });
    
    setNewExtension("");
    setNewCommand("");
  };

  const handleRemoveExtension = (extension: string) => {
    if (!config) return;
    
    const { [extension]: _, ...rest } = config.extensionCommands;
    
    updateConfig({
      ...config,
      extensionCommands: rest
    });
  };

  const handleUpdateExtensionCommand = (extension: string, command: string) => {
    if (!config) return;
    
    updateConfig({
      ...config,
      extensionCommands: {
        ...config.extensionCommands,
        [extension]: command
      }
    });
  };

  const handleReloadScripts = async () => {
    try {
      await reloadScripts();
      toast({
        title: "Success",
        description: "Scripts reloaded successfully",
      });
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to reload scripts",
        variant: "destructive",
      });
    }
  };

  const EnvVariablesForm = () => {
    const form = useForm({
      defaultValues: {
        key: "",
        value: ""
      }
    });

    const handleAddEnvVar = (data: { key: string, value: string }) => {
      if (!config || !data.key) return;

      updateConfig({
        ...config,
        environmentVariables: {
          ...config.environmentVariables,
          [data.key]: data.value
        }
      });

      form.reset();
    };

    const handleDeleteEnvVar = (key: string) => {
      if (!config) return;

      const { [key]: _, ...rest } = config.environmentVariables || {};
      
      updateConfig({
        ...config,
        environmentVariables: rest
      });
    };

    return (
      <div className="space-y-4">
        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleAddEnvVar)} className="space-y-4">
            <FormField
              control={form.control}
              name="key"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Key</FormLabel>
                  <FormControl>
                    <Input placeholder="Enter environment variable key" {...field} />
                  </FormControl>
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="value"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Value</FormLabel>
                  <FormControl>
                    <Input placeholder="Enter environment variable value" {...field} />
                  </FormControl>
                </FormItem>
              )}
            />
            <Button type="submit">Add Variable</Button>
          </form>
        </Form>

        <div className="space-y-2">
          <h4 className="text-sm font-medium">Current Variables</h4>
          {config?.environmentVariables && Object.entries(config.environmentVariables).map(([key, value]) => (
            <div key={key} className="flex items-center justify-between p-2 border rounded">
              <div>
                <span className="font-mono">{key}</span>
                <span className="text-muted-foreground ml-2">=</span>
                <span className="font-mono ml-2">{value}</span>
              </div>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => handleDeleteEnvVar(key)}
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          ))}
        </div>
      </div>
    );
  };

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (!config) {
    return <div>Error loading configuration</div>;
  }

  return (
    <AppLayout>
      <div className="container py-6 space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold">Settings</h1>
          <div className="flex items-center gap-2">
            <Button variant="outline" onClick={handleReloadScripts}>
              <RefreshCw className="h-4 w-4 mr-2" />
              Reload Scripts
            </Button>
            <Button onClick={handleSave}>
              <Save className="h-4 w-4 mr-2" />
              Save Changes
            </Button>
          </div>
        </div>

        <Tabs defaultValue="general">
          <TabsList>
            <TabsTrigger value="general">General</TabsTrigger>
            <TabsTrigger value="extensions">Extension Commands</TabsTrigger>
            <TabsTrigger value="env">Environment Variables</TabsTrigger>
          </TabsList>

          <TabsContent value="general">
            <Card>
              <CardHeader>
                <CardTitle>General Settings</CardTitle>
                <CardDescription>
                  Configure general application settings
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-6">
                  <div className="space-y-4">
                    <div>
                      <Label htmlFor="editor">Default Editor</Label>
                      <Input
                        id="editor"
                        placeholder="Enter default editor command"
                        value={config.editor || ''}
                        onChange={(e) => updateConfig({
                          ...config,
                          editor: e.target.value
                        })}
                      />
                    </div>
                  </div>
                  <Separator />
                  <div className="space-y-4">
                    <div className="flex gap-2">
                      <Input
                        placeholder="Enter folder path"
                        value={newFolder}
                        onChange={(e) => setNewFolder(e.target.value)}
                      />
                      <Button onClick={handleAddFolder}>
                        <Plus className="h-4 w-4 mr-2" />
                        Add Folder
                      </Button>
                    </div>
                    <div className="space-y-2">
                      {config.scriptFolders.map((folder) => (
                        <div key={folder} className="flex items-center justify-between p-2 border rounded">
                          <div className="flex items-center gap-2">
                            <FolderOpen className="h-4 w-4 text-muted-foreground" />
                            <span className="font-mono">{folder}</span>
                          </div>
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => handleRemoveFolder(folder)}
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="extensions">
            <Card>
              <CardHeader>
                <CardTitle>Extension Commands</CardTitle>
                <CardDescription>
                  Configure commands for different file extensions
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div className="flex gap-2">
                    <Input
                      placeholder="Extension (e.g. .py)"
                      value={newExtension}
                      onChange={(e) => setNewExtension(e.target.value)}
                    />
                    <Input
                      placeholder="Command (e.g. python)"
                      value={newCommand}
                      onChange={(e) => setNewCommand(e.target.value)}
                    />
                    <Button onClick={handleAddExtension}>
                      <Plus className="h-4 w-4 mr-2" />
                      Add Extension
                    </Button>
                  </div>
                  <div className="space-y-2">
                    {Object.entries(config.extensionCommands).map(([extension, command]) => (
                      <div key={extension} className="flex items-center gap-2">
                        <Badge variant="outline">{extension}</Badge>
                        <Input
                          value={command}
                          onChange={(e) => handleUpdateExtensionCommand(extension, e.target.value)}
                        />
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleRemoveExtension(extension)}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    ))}
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="env">
            <Card>
              <CardHeader>
                <CardTitle>Environment Variables</CardTitle>
                <CardDescription>
                  Add or remove environment variables for script execution
                </CardDescription>
              </CardHeader>
              <CardContent>
                <EnvVariablesForm />
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </AppLayout>
  );
};

export default Settings;
