import { useState } from "react";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { navigationService } from "@/services/navigationService";

interface ApiKeyModalProps {
  isOpen: boolean;
}

export function ApiKeyModal({ isOpen}: ApiKeyModalProps) {
  const [apiKey, setApiKey] = useState("");

  const handleSave = () => {
    if (apiKey.trim()) {
      localStorage.setItem('dev-loop-api-key', apiKey);
      setApiKey("");
      navigationService.toHome();
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={() => navigationService.toHome()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>API Key Required</DialogTitle>
          <DialogDescription>
            Please enter your API key to continue. This will be stored in your browser's localStorage.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="apiKey">API Key</Label>
            <Input
              id="apiKey"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              placeholder="Enter your API key"
            />
          </div>
        </div>
        <DialogFooter>
          <Button onClick={handleSave}>Save</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
} 