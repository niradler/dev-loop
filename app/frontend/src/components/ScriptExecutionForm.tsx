import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Script, ScriptInput, ScriptExecutionInput } from "@/types/script";
import { PlayCircle, EyeOff } from "lucide-react";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

interface ScriptExecutionFormProps {
  script: Script;
  onExecute: (inputs: ScriptExecutionInput[], isIncognito: boolean) => Promise<void>;
  isExecuting: boolean;
}

export function ScriptExecutionForm({ script, onExecute, isExecuting }: ScriptExecutionFormProps) {
  const [inputs, setInputs] = useState<ScriptExecutionInput[]>([]);
  const [isIncognito, setIsIncognito] = useState(false);

  useEffect(() => {
    if (script?.inputs) {
      setInputs(script.inputs.map(input => ({
        name: input.name,
        value: input.default || ''
      })));
    }
  }, [script?.inputs]);

  const handleInputChange = (index: number, value: string | number | boolean) => {
    setInputs(prev => prev.map((input, i) => 
      i === index ? { ...input, value } : input
    ));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    await onExecute(inputs, isIncognito);
  };

  const renderInput = (input: ScriptInput, index: number) => {
    if (!input) return null;

    switch (input.type) {
      case 'boolean':
        return (
          <div className="flex items-center space-x-2">
            <Switch
              id={input.name}
              checked={Boolean(inputs[index]?.value)}
              onCheckedChange={(checked) => handleInputChange(index, checked)}
            />
            <Label htmlFor={input.name}>{input.name}</Label>
          </div>
        );
      
      case 'number':
        return (
          <Input
            type="number"
            id={input.name}
            placeholder={input.description}
            value={inputs[index]?.value as number || ''}
            onChange={(e) => handleInputChange(index, Number(e.target.value))}
            required={input.required}
          />
        );
      
      case 'select':
        return (
          <Select 
            value={String(inputs[index]?.value || '')} 
            onValueChange={(value) => handleInputChange(index, value)}
          >
            <SelectTrigger>
              <SelectValue placeholder={input.description} />
            </SelectTrigger>
            <SelectContent>
              {input.options?.map((option) => (
                <SelectItem key={option} value={option}>
                  {option}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        );
      
      case 'string':
      default:
        return (
          <Input
            type="text"
            id={input.name}
            placeholder={input.description}
            value={inputs[index]?.value as string || ''}
            onChange={(e) => handleInputChange(index, e.target.value)}
            required={input.required}
          />
        );
    }
  };

  if (!script?.inputs) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Execute Script</CardTitle>
          <CardDescription>
            No inputs required for this script.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center space-x-2">
            <Switch
              id="incognito"
              checked={isIncognito}
              onCheckedChange={setIsIncognito}
            />
            <Label htmlFor="incognito" className="flex items-center gap-2">
              <EyeOff className="h-4 w-4" />
              Incognito Mode
            </Label>
          </div>
        </CardContent>
        <CardFooter>
          <Button onClick={() => onExecute([], isIncognito)} disabled={isExecuting}>
            {isExecuting ? 'Executing...' : 'Execute'}
          </Button>
        </CardFooter>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Execute Script</CardTitle>
        <CardDescription>
          Fill in the required inputs and click Execute to run the script.
        </CardDescription>
      </CardHeader>
      <form onSubmit={handleSubmit}>
        <CardContent className="space-y-4">
          {script.inputs.map((input, index) => (
            <div key={input.name} className="space-y-2">
              <Label htmlFor={input.name}>
                {input.name}
                {input.required && <span className="text-destructive ml-1">*</span>}
              </Label>
              {renderInput(input, index)}
            </div>
          ))}
          <div className="flex items-center space-x-2">
            <Switch
              id="incognito"
              checked={isIncognito}
              onCheckedChange={setIsIncognito}
            />
            <Label htmlFor="incognito" className="flex items-center gap-2">
              <EyeOff className="h-4 w-4" />
              Incognito Mode
            </Label>
          </div>
        </CardContent>
        <CardFooter>
          <Button type="submit" disabled={isExecuting}>
            {isExecuting ? 'Executing...' : 'Execute'}
          </Button>
        </CardFooter>
      </form>
    </Card>
  );
}
